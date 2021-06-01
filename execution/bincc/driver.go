// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package bincc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/aungmawjj/juria-blockchain/execution/chaincode"
	"golang.org/x/crypto/sha3"
)

type CodeDriver struct {
	codeDir     string
	execTimeout time.Duration
	mtxInstall  sync.Mutex
}

func NewCodeDriver(codeDir string, timeout time.Duration) *CodeDriver {
	return &CodeDriver{
		codeDir:     codeDir,
		execTimeout: timeout,
	}
}

func (drv *CodeDriver) Install(codeID, data []byte) error {
	drv.mtxInstall.Lock()
	defer drv.mtxInstall.Unlock()
	return drv.downloadCodeIfRequired(codeID, data)
}

func (drv *CodeDriver) GetInstance(codeID []byte) (chaincode.Chaincode, error) {
	return &Runner{
		codePath: path.Join(drv.codeDir, hex.EncodeToString(codeID)),
		timeout:  drv.execTimeout,
	}, nil
}

func (drv *CodeDriver) downloadCodeIfRequired(codeID, data []byte) error {
	filepath := path.Join(drv.codeDir, hex.EncodeToString(codeID))
	if _, err := os.Stat(filepath); err == nil {
		return nil // code file already exist
	}
	resp, err := drv.downloadCode(string(data), 5)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	sum, buf, err := copyAndSumCode(resp.Body)
	if err != nil {
		return err
	}
	if !bytes.Equal(codeID, sum) {
		return fmt.Errorf("invalid code hash")
	}
	return writeCodeFile(drv.codeDir, codeID, buf)
}

func (drv *CodeDriver) downloadCode(url string, retry int) (*http.Response, error) {
	resp, err := http.Get(url)
	if err == nil {
		if resp.StatusCode != 200 {
			err = fmt.Errorf("status not 200")
		}
	}
	if err != nil {
		if retry <= 0 {
			return nil, err
		}
		time.Sleep(100 * time.Millisecond)
		return drv.downloadCode(url, retry-1)
	}
	return resp, nil
}

func writeCodeFile(codeDir string, codeID []byte, r io.Reader) error {
	filename := hex.EncodeToString(codeID)
	filepath := path.Join(codeDir, filename)
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return err
	}
	return os.Chmod(filepath, 0755)
}

func copyAndSumCode(r io.Reader) ([]byte, *bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	h := sha3.New256()
	if _, err := io.Copy(buf, io.TeeReader(r, h)); err != nil {
		return nil, nil, err
	}
	return h.Sum(nil), buf, nil
}

func StoreCode(codeDir string, r io.Reader) ([]byte, error) {
	codeID, buf, err := copyAndSumCode(r)
	if err != nil {
		return nil, err
	}
	if err := writeCodeFile(codeDir, codeID, buf); err != nil {
		return nil, err
	}
	return codeID, nil
}
