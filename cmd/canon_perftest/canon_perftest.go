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

package ginja


const char kPath[] =
    "../../third_party/WebKit/Source/WebCore/platform/leveldb/LevelDBWriteBatch.cpp"

func main() int {
  var times []int

  char buf[200]
  len2 := strlen(kPath)
  strcpy(buf, kPath)

  for j := 0; j < 5; j++ {
    kNumRepetitions := 2000000
    start := GetTimeMillis()
    var slash_bits uint64
    for i := 0; i < kNumRepetitions; i++ {
      CanonicalizePath(buf, &len2, &slash_bits)
    }
    int delta = (int)(GetTimeMillis() - start)
    times.push_back(delta)
  }

  min := times[0]
  max := times[0]
  total := 0
  for i := 0; i < times.size(); i++ {
    total += times[i]
    if times[i] < min {
      min = times[i]
    } else if times[i] > max {
      max = times[i]
    }
  }

  printf("min %dms  max %dms  avg %.1fms\n", min, max, total / times.size())
}
