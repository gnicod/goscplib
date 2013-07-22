package goscplib

import(
	"code.google.com/p/go.crypto/ssh"
	"fmt"
	"log"
	"os"
	"io"
	"path/filepath"
)

//Constants
const (
	SCP_PUSH_BEGIN_FILE   = "C0644"
	SCP_PUSH_BEGIN_FOLDER = "D0755"
	SCP_PUSH_END          = "\x00"
)

type Scp struct {
	client *ssh.ClientConn
}

//Initializer
func NewScp(clientConn *ssh.ClientConn ) *Scp{
	return &Scp{
		client: clientConn,
	}
}

//Push one file to server
func (scp *Scp) PushFile (src string, dest string) error{
	session, err := scp.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fileSrc, srcErr := os.Open(src)
		if srcErr != nil {
			log.Fatalln("Failed to open source file: " + srcErr.Error())
		}
		//Get file size
		srcStat, statErr := fileSrc.Stat()
		if statErr != nil {
			log.Fatalln("Failed to stat file: " + statErr.Error())
		}
		// According to https://blogs.oracle.com/janp/entry/how_the_scp_protocol_works
		// Print the file content
		fmt.Fprintln(w, SCP_PUSH_BEGIN_FILE, srcStat.Size(), filepath.Base(dest))
		io.Copy(w, fileSrc)
		fmt.Fprint(w, SCP_PUSH_END)
	}()
	if err := session.Run("/usr/bin/scp -qrt "+filepath.Dir(dest)); err != nil {
		return err
	}
	return nil
}

//Push directory to server
func (scp *Scp) PushDir (src string, dest string) {
	fmt.Println("not implemented yet")
}
