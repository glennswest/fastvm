package main

import (
    "bufio"
    "fmt"
    "os"
    "log"
    "math/rand"
    "time"
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
