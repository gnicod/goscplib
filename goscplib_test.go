package goscplib

import (
	"testing"

	"golang.org/x/crypto/ssh"
)

func newClient() (*ssh.Client, error) {
	return NewClient("192.168.2.40", "root", "redhat")
}

func TestPushFile(t *testing.T) {
	client, err := newClient()
	if err != nil {
		t.Fatal(err)
	}
	scp := NewScp(client)
	srcFile := "/home/fugr/gocode/src/github.com/upmio/mgserver/cleanup/cleanup"
	dest := "/root/testdir/cleanup"
	err = scp.PushFile(srcFile, dest)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPushDir(t *testing.T) {
	client, err := newClient()
	if err != nil {
		t.Fatal(err)
	}
	scp := NewScp(client)
	srcDir := "/home/fugr/gocode/src/github.com/upmio/mgserver/models/"
	dest := "/root/testdir/"
	err = scp.PushDir(srcDir, dest)
	if err != nil {
		t.Fatal(err)
	}
}
