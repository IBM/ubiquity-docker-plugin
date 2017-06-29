package block_device_utils

import (
    "github.com/IBM/ubiquity/utils/logs"
    "path"
)

const bindSuffix = "ub_root"


func (s *impBlockDeviceUtils) MountFs(mpath string, mpoint string) error {
    defer s.logger.Trace(logs.DEBUG)()

    // verify cmd exists
    mountCmd := "mount"
    if err := s.exec.IsExecutable(mountCmd); err != nil {
        return s.logger.ErrorRet(&commandNotFoundError{mountCmd, err}, "failed")
    }

    // create the mount point to the fs root
    mpointFsRoot := path.Join("/var", "run", "ubiquity", path.Base(mpoint))
    if err := s.exec.MkdirAll(mpointFsRoot, 0700); err != nil {
        return s.logger.ErrorRet(err, "MkdirAll failed", logs.Args{{"mpointFsRoot", mpointFsRoot}})
    }

    // mount mpointFsRoot
    args := []string{mountCmd, mpath, mpointFsRoot}
    if _, err := s.exec.Execute("sudo", args); err != nil {
        return s.logger.ErrorRet(&commandExecuteError{mountCmd, err}, "failed")
    }

    // create bind dir
    bindDir := path.Join(mpointFsRoot, bindSuffix)
    if err := s.exec.MkdirAll(bindDir, 0700); err != nil {
        return s.logger.ErrorRet(err, "MkdirAll failed", logs.Args{{"bindDir", bindDir}})
    }

    // mount bind dir
    args = []string{mountCmd, "--bind", bindDir, mpoint}
    if _, err := s.exec.Execute("sudo", args); err != nil {
        return s.logger.ErrorRet(&commandExecuteError{mountCmd, err}, "failed")
    }

    s.logger.Info("mounted", logs.Args{{"mpoint", mpoint}})
    return nil
}

func (s *impBlockDeviceUtils) UmountFs(mpoint string) error {
    defer s.logger.Trace(logs.DEBUG)()

    // verify cmd exists
    umountCmd := "umount"
    if err := s.exec.IsExecutable(umountCmd); err != nil {
        return s.logger.ErrorRet(&commandNotFoundError{umountCmd, err}, "failed")
    }

    // umount twice - the bind dir and device
    args := []string{umountCmd, mpoint}
    if _, err := s.exec.Execute("sudo", args); err != nil {
        return s.logger.ErrorRet(&commandExecuteError{umountCmd, err}, "failed")
    }
    if _, err := s.exec.Execute("sudo", args); err != nil {
        return s.logger.ErrorRet(&commandExecuteError{umountCmd, err}, "failed")
    }

    s.logger.Info("umounted", logs.Args{{"mpoint", mpoint}})
    return nil
}
