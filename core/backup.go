package core

import (
	"ach/bootstrap"
	"ach/pkg/utils"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"sort"
	"time"
)

type Backup struct {
	FileName string
	Size     int64
	modTime  time.Time
}

type Snapshot struct {
	Backup
}

func (backup Backup) ModTime() string {
	return utils.FormatTime(backup.modTime)
}

func NewBackup(fileName string, size int64, modTime time.Time) Backup {
	return Backup{FileName: fileName, Size: size, modTime: modTime}
}

func (server *Server) GetBackupList(dst string) []Backup {
	backupList := make([]Backup, 0, 10)

	res, _ := ioutil.ReadDir(dst)
	sort.Slice(res, func(i, j int) bool {
		return res[i].ModTime().Unix() < res[j].ModTime().Unix()
	})
	for _, f := range res {
		backupList = append(backupList, NewBackup(f.Name(), f.Size(), f.ModTime()))
	}
	return backupList
}

func (server *Server) MakeBackup(dst string) error {
	src := path.Join(filepath.Dir(server.config.ExecPath), "world")
	log.Printf("[%s/INFO]: Making backup to %s...\n", server.Name, dst)
	server.Write(fmt.Sprintf("say Making backup to %s...", dst))

	server.Write("save-off")
	err := utils.CopyDir(src, dst)
	server.Write("save-on")

	if err != nil {
		log.Printf("[%s/ERROR]: Backup making failed.\n", server.Name)
		server.Write("say Backup making failed.")
		return err
	}
	log.Printf("[%s/INFO]: Backup making succeed.\n", server.Name)
	server.Write("say Backup making succeed.")
	return nil
}

func (server *Server) LoadBackup(src string) error {
	dst := path.Join(filepath.Dir(server.config.ExecPath), "world")
	server.MakeSnapshot("Before loading backup")

	server.Write("stop")
	for server.Running {
		time.Sleep(time.Second)
	}

	utils.DeletePath(dst)

	log.Printf("[%s/INFO]: Loading backup %s...\n", server.Name, src)
	err := utils.CopyDir(src, dst)
	if err != nil {
		log.Printf("[%s/ERROR]: Backup loading failed.\n", server.Name)
		return err
	}
	log.Printf("[%s/INFO]: Backup loaded successfully.\n", server.Name)

	server.start()
	go server.wait()

	return nil
}

func (server *Server) GetSnapshotCount() int {
	res, _ := ioutil.ReadDir(path.Join(bootstrap.Config.BackupDir, "snapshots"))
	return len(res)
}

func (server *Server) GetSnapshotList() []Backup {
	return server.GetBackupList(path.Join(bootstrap.Config.BackupDir, "snapshots"))
}

func (server *Server) DeleteOldSnapshot() {
	res, _ := ioutil.ReadDir(path.Join(bootstrap.Config.BackupDir, "snapshots"))
	sort.Slice(res, func(i, j int) bool {
		return res[i].ModTime().Unix() < res[j].ModTime().Unix()
	})

	utils.DeletePath(path.Join(bootstrap.Config.BackupDir, "snapshots", res[0].Name()))
}

func (server *Server) MakeSnapshot(comment string) error {
	for server.GetSnapshotCount() >= 10 {
		server.DeleteOldSnapshot()
	}

	bkname := utils.GetTimeStamp()
	if len(comment) > 0 {
		bkname += " " + comment
	}

	if err := server.MakeBackup(
		path.Join(
			bootstrap.Config.BackupDir,
			"snapshots",
			fmt.Sprintf("%s - %s", server.Name, bkname),
		),
	); err != nil {
		return err
	}
	return nil
}

func (server *Server) LoadSnapshot(index int) error {
	snapshotName := server.GetSnapshotList()[index].FileName
	snapshotPath := path.Join(bootstrap.Config.BackupDir, "snapshots", snapshotName)
	if err := server.LoadBackup(snapshotPath); err != nil {
		return err
	}
	return nil
}
