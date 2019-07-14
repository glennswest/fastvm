package main

import (
        //"io"
        "io/ioutil"
        //"net"
        //"github.com/tidwall/sjson"
        //"github.com/tidwall/gjson"
        . "github.com/glennswest/go-sshclient"
        b64 "encoding/base64"
        "strings"
         "os"
         //"encoding/json"
         "fmt"
         "log"
         "errors"
         "time"
         "bufio"
         "math/rand"
         bolt "go.etcd.io/bbolt"
)


func main(){
    vmname := "rat.swamp.k.lo"
    mac := make_mac()
    make_vmx(vmname,mac)
    make_pxefile(vmname,mac)

}

func make_pxefile(vname string,mac string){
    osimage := "http://svcrhcos.app.ctl.k.e2e.bos.redhat.com/rhcos-bios.raw.gz"
    ignition := "http://svcrhcos.app.ctl.k.e2e.bos.redhat.com/tiny/bootstrap.ign"
    os.MkdirAll("/var/lib/tftpboot/pxelinux/pxelinux.cfg",777)
    f, _ := os.Create(mac)
    w := bufio.NewWriter(f)
    fmt.Fprintln(w,"# %s",vname)
    fmt.Fprintln(w,"default coreos")
    fmt.Fprintln(w,"prompt 0")
    fmt.Fprintln(w,"timeout 1")
    fmt.Fprintln(w,"")
    fmt.Fprintln(w,"")
    fmt.Fprintln(w,"label coreos")
    fmt.Fprintln(w,"  kernel rhcosinstall.vmlinuz")
    fmt.Fprint(w,"  append ip=ens192:dhcp rd.neednet=1 initrd=rhcosinstall-initramfs.img coreos.inst=yes coreos.inst.install_dev=sda coreos.inst.image_url=")
    fmt.Fprint(w,"%s",osimage)
    fmt.Fprint(w," coreos.inst.ignition_url=")
    fmt.Fprint(w,"%s\n",ignition)
    w.Flush()
}

func make_mac() string {
    mac := "00:50:56";
    rand.Seed(time.Now().UnixNano())
    token := make([]byte, 3)
    rand.Read(token)
    macx := fmt.Sprintf(":%2.2x:%2.2x:%2.2x", token[0], token[1], token[2])
    mac = mac + macx
    log.Printf(mac);
    return(mac)
}

func esxi_command(cmd string){

    host := GetDbValue("esxihost")
    log.Printf("esxi host: %s\n",host)
    username := GetDbValue("sshuser")
    sshkey_path := "~/.fastvm"
    os.MkdirAll(sshkey_path,0700)
    sshkey_path = sshkey_path + "/id"
    sshkeyb64 := GetDbValue("sshkey")
    sshkeybytes, _ := b64.StdEncoding.DecodeString(sshkeyb64)
    ioutil.WriteFile(sshkey_path, sshkeybytes, 0600)
    SshCommand(host,username,sshkey_path,cmd)
}
 
func SshCommand(host string,username string,keypath string,cmd string) string{

    hp := host
    if (strings.ContainsAny(hp,":") == false){
       hp = host + ":22"
       }
    // Keypath should be pathname of private key
    client, err := DialWithKey(hp, username, keypath)
    if (err != nil){
       log.Printf("SSHCommand: Cannot Connect to %s - %v\n",host,err)
       return("")
       }
    defer client.Close()
    log.Printf("%s\n",cmd)
    out, err := client.Cmd(cmd).SmartOutput()
    if (err != nil){
          log.Printf("SSHCommand: Cannot send cmd: %v\n",err)
          }
    log.Printf("ssh: %s\n",string(out))
    return(string(out))
}


func make_vmx(name string,mac string){
    f, _ := os.Create(name+".vmx")
    w := bufio.NewWriter(f)
    fmt.Fprintln(w,`config.version = "8"`)
    fmt.Fprintln(w,`virtualHW.version = "14"`)
    fmt.Fprintln(w,`memSize = "16384"`)
    fmt.Fprintln(w,`bios.bootRetry.delay = "10"`)
    fmt.Fprintln(w,`scsi0.virtualDev = "lsisas1068"`)
    fmt.Fprintln(w,`scsi0.present = "TRUE"`)
    fmt.Fprintln(w,`scsi0:0.deviceType = "scsi-hardDisk"`)
    thename := "scsi0:0.fileName = \"" + name + ".vmdk\""
    fmt.Fprintln(w,thename)
    fmt.Fprintln(w,`scsi0:0.present = "TRUE"`)
    fmt.Fprintln(w,`ethernet0.virtualDev = "e1000"`)
    fmt.Fprintln(w,`ethernet0.networkName = "Glan Network"`)
    fmt.Fprintln(w,`ethernet0.addressType = "static"`)
    fmt.Fprintln(w,`ethernet0.present = "TRUE"`)
    fmt.Fprintln(w,"displayName = " + "\"" + name + "\"")
    fmt.Fprintln(w,`guestOS = "coreos-64"`)
    fmt.Fprintln(w,`bios.bootRetry.enabled = "TRUE"`)
    fmt.Fprintln(w,"nvram = \"" + name + ".nvram\"")
    fmt.Fprintln(w,"ethernet0.address = \"" + mac + "\"")
    w.Flush()
}
// Exists reports whether the named file or directory exists.
func Exists(name string) bool {
    if _, err := os.Stat(name); err != nil {
        if os.IsNotExist(err) {
            return false
        }
    }
    return true
}

func GetBucketAndKey(k string) (string, string){
     idx := strings.IndexByte(k,'.')
     bucket := k[0:idx]
     key := k[idx+1:]
     return bucket,key
}

func GetDbValue(p string) string{
     val := ""
     //log.Printf("GetDbValue(%s)\n",p)
     b,k := GetBucketAndKey(p)
     db, _ := bolt.Open("/data/winoperator", 0600,nil)
     db.View(func(tx *bolt.Tx) error {
              bucket := tx.Bucket([]byte(b))
              if bucket == nil {
                   log.Printf("No Bucket\n")
                   return errors.New("No Key ")
                   }

        val = string(bucket.Get([]byte(k)))
        return nil
    })
     //log.Printf("External value = %s\n",val)
     db.Close()
     return val
}
func SetDbValue(k string,v string){
     //log.Printf("SetDbValue(%s,%s)\n",k,v)
     bucket,key := GetBucketAndKey(k)
     db, _ := bolt.Open("/data/winoperator", 0600,nil)
     db.Update(func(tx *bolt.Tx) error {
           b, err := tx.CreateBucketIfNotExists([]byte(bucket))
           if err != nil {
              log.Printf("Err in create of bucket\n")
              return err
              }
           //log.Printf("Setting %s to %s\n",key,v)
           b.Put([]byte(key),[]byte(v))
           return(nil)
           })
     db.Close()
     return
}

