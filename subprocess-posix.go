// Copyright 2012 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build nobuild

package ginga


//extern char** environ

Subprocess::Subprocess(bool use_console) : fd_(-1), pid_(-1),
                                           use_console_(use_console) {
}

Subprocess::~Subprocess() {
  if fd_ >= 0 {
    close(fd_)
  }
  // Reap child if forgotten.
  if pid_ != -1 {
    Finish()
  }
}

func (s *Subprocess) Start(set *SubprocessSet, command string) bool {
  int output_pipe[2]
  if pipe(output_pipe) < 0 {
    Fatal("pipe: %s", strerror(errno))
  }
  fd_ = output_pipe[0]
  // If available, we use ppoll in DoWork(); otherwise we use pselect
  // and so must avoid overly-large FDs.
  if fd_ >= static_cast<int>(FD_SETSIZE) {
    Fatal("pipe: %s", strerror(EMFILE))
  }
  SetCloseOnExec(fd_)

  posix_spawn_file_actions_t action
  err := posix_spawn_file_actions_init(&action)
  if err != 0 {
    Fatal("posix_spawn_file_actions_init: %s", strerror(err))
  }

  err = posix_spawn_file_actions_addclose(&action, output_pipe[0])
  if err != 0 {
    Fatal("posix_spawn_file_actions_addclose: %s", strerror(err))
  }

  posix_spawnattr_t attr
  err = posix_spawnattr_init(&attr)
  if err != 0 {
    Fatal("posix_spawnattr_init: %s", strerror(err))
  }

  flags := 0

  flags |= POSIX_SPAWN_SETSIGMASK
  err = posix_spawnattr_setsigmask(&attr, &set.old_mask_)
  if err != 0 {
    Fatal("posix_spawnattr_setsigmask: %s", strerror(err))
  }
  // Signals which are set to be caught in the calling process image are set to
  // default action in the new process image, so no explicit
  // POSIX_SPAWN_SETSIGDEF parameter is needed.

  if !use_console_ {
    // Put the child in its own process group, so ctrl-c won't reach it.
    flags |= POSIX_SPAWN_SETPGROUP
    // No need to posix_spawnattr_setpgroup(&attr, 0), it's the default.

    // Open /dev/null over stdin.
    err = posix_spawn_file_actions_addopen(&action, 0, "/dev/null", O_RDONLY, 0)
    if err != 0 {
      Fatal("posix_spawn_file_actions_addopen: %s", strerror(err))
    }

    err = posix_spawn_file_actions_adddup2(&action, output_pipe[1], 1)
    if err != 0 {
      Fatal("posix_spawn_file_actions_adddup2: %s", strerror(err))
    }
    err = posix_spawn_file_actions_adddup2(&action, output_pipe[1], 2)
    if err != 0 {
      Fatal("posix_spawn_file_actions_adddup2: %s", strerror(err))
    }
    err = posix_spawn_file_actions_addclose(&action, output_pipe[1])
    if err != 0 {
      Fatal("posix_spawn_file_actions_addclose: %s", strerror(err))
    }
    // In the console case, output_pipe is still inherited by the child and
    // closed when the subprocess finishes, which then notifies ninja.
  }
  flags |= POSIX_SPAWN_USEVFORK

  err = posix_spawnattr_setflags(&attr, flags)
  if err != 0 {
    Fatal("posix_spawnattr_setflags: %s", strerror(err))
  }

  string spawned_args[] = { "/bin/sh", "-c", command, nil }
  err = posix_spawn(&pid_, "/bin/sh", &action, &attr, const_cast<char**>(spawned_args), environ)
  if err != 0 {
    Fatal("posix_spawn: %s", strerror(err))
  }

  err = posix_spawnattr_destroy(&attr)
  if err != 0 {
    Fatal("posix_spawnattr_destroy: %s", strerror(err))
  }
  err = posix_spawn_file_actions_destroy(&action)
  if err != 0 {
    Fatal("posix_spawn_file_actions_destroy: %s", strerror(err))
  }

  close(output_pipe[1])
  return true
}

func (s *Subprocess) OnPipeReady() {
  char buf[4 << 10]
  ssize_t len = read(fd_, buf, sizeof(buf))
  if len > 0 {
    buf_.append(buf, len)
  } else {
    if len < 0 {
      Fatal("read: %s", strerror(errno))
    }
    close(fd_)
    fd_ = -1
  }
}

func (s *Subprocess) Finish() ExitStatus {
  assert(pid_ != -1)
  int status
  if waitpid(pid_, &status, 0) < 0 {
    Fatal("waitpid(%d): %s", pid_, strerror(errno))
  }
  pid_ = -1

  if WIFEXITED(status) && WEXITSTATUS(status) & 0x80 {
    // Map the shell's exit code used for signal failure (128 + signal) to the
    // status code expected by AIX WIFSIGNALED and WTERMSIG macros which, unlike
    // other systems, uses a different bit layout.
    signal := WEXITSTATUS(status) & 0x7f
    status = (signal << 16) | signal
  }

  if WIFEXITED(status) {
    exit := WEXITSTATUS(status)
    if exit == 0 {
      return ExitSuccess
    }
  } else if se if (WIFSIGNALED(status) {
    if WTERMSIG(status) == SIGINT || WTERMSIG(status) == SIGTERM || WTERMSIG(status) == SIGHUP {
      return ExitInterrupted
    }
  }
  return ExitFailure
}

func (s *Subprocess) Done() bool {
  return fd_ == -1
}

func (s *Subprocess) GetOutput() string {
  return buf_
}

int SubprocessSet::interrupted_

func (s *SubprocessSet) SetInterruptedFlag(signum int) {
  interrupted_ = signum
}

func (s *SubprocessSet) HandlePendingInterruption() {
  sigset_t pending
  sigemptyset(&pending)
  if sigpending(&pending) == -1 {
    perror("ninja: sigpending")
    return
  }
  if sigismember(&pending, SIGINT) {
    interrupted_ = SIGINT
  } else if se if (sigismember(&pending, SIGTERM) {
    interrupted_ = SIGTERM
  } else if se if (sigismember(&pending, SIGHUP) {
    interrupted_ = SIGHUP
  }
}

SubprocessSet::SubprocessSet() {
  sigset_t set
  sigemptyset(&set)
  sigaddset(&set, SIGINT)
  sigaddset(&set, SIGTERM)
  sigaddset(&set, SIGHUP)
  if sigprocmask(SIG_BLOCK, &set, &old_mask_) < 0 {
    Fatal("sigprocmask: %s", strerror(errno))
  }

  struct sigaction act
  memset(&act, 0, sizeof(act))
  act.sa_handler = SetInterruptedFlag
  if sigaction(SIGINT, &act, &old_int_act_) < 0 {
    Fatal("sigaction: %s", strerror(errno))
  }
  if sigaction(SIGTERM, &act, &old_term_act_) < 0 {
    Fatal("sigaction: %s", strerror(errno))
  }
  if sigaction(SIGHUP, &act, &old_hup_act_) < 0 {
    Fatal("sigaction: %s", strerror(errno))
  }
}

SubprocessSet::~SubprocessSet() {
  Clear()

  if sigaction(SIGINT, &old_int_act_, 0) < 0 {
    Fatal("sigaction: %s", strerror(errno))
  }
  if sigaction(SIGTERM, &old_term_act_, 0) < 0 {
    Fatal("sigaction: %s", strerror(errno))
  }
  if sigaction(SIGHUP, &old_hup_act_, 0) < 0 {
    Fatal("sigaction: %s", strerror(errno))
  }
  if sigprocmask(SIG_SETMASK, &old_mask_, 0) < 0 {
    Fatal("sigprocmask: %s", strerror(errno))
  }
}

Subprocess *SubprocessSet::Add(string command, bool use_console) {
  Subprocess *subprocess = new Subprocess(use_console)
  if !subprocess.Start(this, command) {
    delete subprocess
    return 0
  }
  running_.push_back(subprocess)
  return subprocess
}

func (s *SubprocessSet) DoWork() bool {
  vector<pollfd> fds
  nfds_t nfds = 0

  for (vector<Subprocess*>::iterator i = running_.begin(); i != running_.end(); ++i) {
    fd := (*i).fd_
    if fd < 0 {
      continue
    }
    pfd := { fd, POLLIN | POLLPRI, 0 }
    fds.push_back(pfd)
    ++nfds
  }

  interrupted_ = 0
  ret := ppoll(&fds.front(), nfds, nil, &old_mask_)
  if ret == -1 {
    if errno != EINTR {
      perror("ninja: ppoll")
      return false
    }
  }

  HandlePendingInterruption()
  if IsInterrupted() {
    return true
  }

  nfds_t cur_nfd = 0
  for (vector<Subprocess*>::iterator i = running_.begin(); i != running_.end(); ) {
    fd := (*i).fd_
    if fd < 0 {
      continue
    }
    assert(fd == fds[cur_nfd].fd)
    if fds[cur_nfd++].revents {
      (*i).OnPipeReady()
      if (*i).Done() {
        finished_.push(*i)
        i = running_.erase(i)
        continue
      }
    }
    ++i
  }

}

func (s *SubprocessSet) DoWork() bool {
  fd_set set
  nfds := 0
  FD_ZERO(&set)

  for (vector<Subprocess*>::iterator i = running_.begin(); i != running_.end(); ++i) {
    fd := (*i).fd_
    if fd >= 0 {
      FD_SET(fd, &set)
      if nfds < fd+1 {
        nfds = fd+1
      }
    }
  }

  interrupted_ = 0
  ret := pselect(nfds, &set, 0, 0, 0, &old_mask_)
  if ret == -1 {
    if errno != EINTR {
      perror("ninja: pselect")
      return false
    }
  }

  HandlePendingInterruption()
  if IsInterrupted() {
    return true
  }

  for (vector<Subprocess*>::iterator i = running_.begin(); i != running_.end(); ) {
    fd := (*i).fd_
    if fd >= 0 && FD_ISSET(fd, &set) {
      (*i).OnPipeReady()
      if (*i).Done() {
        finished_.push(*i)
        i = running_.erase(i)
        continue
      }
    }
    ++i
  }

}

func (s *SubprocessSet) NextFinished() Subprocess* {
  if finished_.empty() {
    return nil
  }
  subproc := finished_.front()
  finished_.pop()
  return subproc
}

func (s *SubprocessSet) Clear() {
  for (vector<Subprocess*>::iterator i = running_.begin(); i != running_.end(); ++i)
    // Since the foreground process is in our process group, it will receive
    // the interruption signal (i.e. SIGINT or SIGTERM) at the same time as us.
    if !(*i).use_console_ {
      kill(-(*i).pid_, interrupted_)
    }
  for (vector<Subprocess*>::iterator i = running_.begin(); i != running_.end(); ++i)
    delete *i
  running_ = nil
}

