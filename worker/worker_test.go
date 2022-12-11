package worker_test

import (
	"testing"

	_ "github.com/Lensual/ServerBackup/backuper/rsyncmd"
	_ "github.com/Lensual/ServerBackup/compressor/zstd"
	"github.com/Lensual/ServerBackup/config"
	_ "github.com/Lensual/ServerBackup/storage/local"
	"github.com/Lensual/ServerBackup/worker"
)

func TestRun(t *testing.T) {
	w := worker.New(
		worker.WithTaskName("mytask"),
		worker.WithTaskConfig(config.TaskTemplate{
			Backup: config.Backup{
				Driver: "rsyncmd",
				Option: map[string]interface{}{
					"rsyncmd": map[string]interface{}{
						"cachePath": "./serverbackup-test/backuper/rsyncmd/cache",
						"host":      "localhost",
						"port":      873,
						"user":      "lensual",
						"useSsh":    true,
						"ssh": map[string]interface{}{
							"port":                  22,
							"privKey":               "~/.ssh/id_rsa",
							"strictHostKeyChecking": false,
						},
						"target":  "/var/www",
						"bwlimit": 0,
						"nice":    19,
						"option": []string{
							"-vv",
						},
					},
				},
			},
			Compress: config.Compress{
				Driver: "zstd",
				Option: map[string]interface{}{},
			},
			Storage: config.Storage{
				Driver: "local",
				Option: map[string]interface{}{
					"local": map[string]interface{}{
						"path": "./serverbackup-test/storage/local/tasknametodo",
					},
				},
			},
		}),
	)
	w.Run()
}
