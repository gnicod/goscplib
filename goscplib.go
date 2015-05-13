package goscplib

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

//Constants
const (
	SCP_PUSH_BEGIN_FILE       = "C"
	SCP_PUSH_BEGIN_FOLDER     = "D"
	SCP_PUSH_BEGIN_END_FOLDER = " 0"
	SCP_PUSH_END_FOLDER       = "E"
	SCP_PUSH_END              = "\x00"
)

func NewClient(addr, user, key string) (*ssh.Client, error) {
	clientConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(key),
		},
	}
	return ssh.Dial("tcp", addr+":22", clientConfig)
}

type Scp struct {
	client *ssh.Client
}

func GetPerm(f *os.File) (perm string) {
	fileStat, _ := f.Stat()
	mod := fileStat.Mode().Perm()
	return fmt.Sprintf("%04o", uint32(mod))
}

//Initializer
func NewScp(client *ssh.Client) *Scp {
	return &Scp{
		client: client,
	}
}

//Push one file to server
func (scp *Scp) PushFile(src string, dest string) error {
	session, err := scp.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		prepareFile(w, src)
	}()
	if err := session.Run("/usr/bin/scp -tr " + filepath.Dir(dest)); err != nil {
		return err
	}
	return nil
}

//Push directory to server
func (scp *Scp) PushDir(src string, dest string) error {
	session, err := scp.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	w, _ := session.StdinPipe()
	//w := os.Stdout
	defer w.Close()
	pushEntries = func(w io.Writer) error {
		folderSrc, err := os.Open(src)
		if err != nil {
			log.Println("Failed to open source file: ", err)
			return
		}
		mode := SCP_PUSH_BEGIN_FOLDER + GetPerm(folderSrc) + SCP_PUSH_BEGIN_END_FOLDER
		fmt.Println(mode)
		fmt.Fprintln(w, mode, filepath.Base(dest))
		lsDir(w, src)
		fmt.Fprintln(w, SCP_PUSH_END_FOLDER)
	}
	if err := session.Run("/usr/bin/scp -qtr" + dest); err != nil {
		return err
	}
	return nil
}

func prepareFile(w io.WriteCloser, src string) error {
	fileSrc, err := os.Open(src)
	//fileStat, err := fileSrc.Stat()
	if err != nil {
		log.Println("Failed to open source file: ", err)
		return err
	}
	defer fileSrc.Close()
	//Get file size
	srcStat, err := fileSrc.Stat()
	if err != nil {
		log.Println("Failed to stat file: ", err)
		return err
	}
	// According to https://blogs.oracle.com/janp/entry/how_the_scp_protocol_works
	// Print the file content
	mode := SCP_PUSH_BEGIN_FILE + GetPerm(fileSrc)
	fmt.Fprintln(w, mode, srcStat.Size(), filepath.Base(src))
	io.Copy(w, fileSrc)
	fmt.Fprint(w, SCP_PUSH_END)
	return nil
}

func lsDir(w io.WriteCloser, dir string) error {
	fi, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	//parcours des dossiers
	for _, f := range fi {
		filename := filepath.Join(dir, f.Name())
		fmt.Println(filename)
		if f.IsDir() {
			folderSrc, err := os.Open(filename)
			if err != nil {
				return err
			}
			defer folderSrc.Close()
			mode := SCP_PUSH_BEGIN_FOLDER + GetPerm(folderSrc)
			fmt.Fprintln(w, mode, SCP_PUSH_BEGIN_END_FOLDER, " ", f.Name())
			lsDir(w, filename)
			fmt.Fprintln(w, SCP_PUSH_END_FOLDER)
		} else {
			if err := prepareFile(w, filename); err != nil {
				return err
			}
		}
	}
	return nil
}
