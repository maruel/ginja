// Copyright 2011 Google Inc. All Rights Reserved.
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

package ginja


func TestStatusTest_StatusFormatElapsed(t *testing.T) {
  BuildConfig config

  status.BuildStarted()
  // Before any task is done, the elapsed time must be zero.
  if "[%/e0.000]" != status.FormatProgressStatus("[%%/e%e]", 0) { t.FailNow() }
}

func TestStatusTest_StatusFormatReplacePlaceholder(t *testing.T) {
  BuildConfig config

  if "[%/s0/t0/r0/u0/f0]" != status.FormatProgressStatus("[%%/s%s/t%t/r%r/u%u/f%f]", 0) { t.FailNow() }
}

