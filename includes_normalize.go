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


// Utility functions for normalizing include paths on Windows.
// TODO: this likely duplicates functionality of CanonicalizePath; refactor.
type IncludesNormalize struct {
  // Normalize path relative to |relative_to|.
  IncludesNormalize(string relative_to)

  // Internal utilities made available for testing, maybe useful otherwise.
  static string AbsPath(StringPiece s, string* err)
  static string Relativize(StringPiece path, const vector<StringPiece>& start_list, string* err)

  string relative_to_
  vector<StringPiece> split_relative_to_
}
