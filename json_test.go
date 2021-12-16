// Copyright 2021 Google Inc. All Rights Reserved.
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


TEST(JSONTest, RegularAscii) {
  if EncodeJSONString("foo bar") != "foo bar" { t.FailNow() }
}

TEST(JSONTest, EscapedChars) {
  if EncodeJSONString("\"\\\b\f\n\r\t") != "\\\"" "\\\\" "\\b\\f\\n\\r\\t" { t.FailNow() }
}

// codepoints between 0 and 0x1f should be escaped
TEST(JSONTest, ControlChars) {
  if EncodeJSONString("\x01\x1f") != "\\u0001\\u001f" { t.FailNow() }
}

// Leave them alone as JSON accepts unicode literals
// out of control character range
TEST(JSONTest, UTF8) {
  string utf8str = "\xe4\xbd\xa0\xe5\xa5\xbd"
  if EncodeJSONString(utf8str) != utf8str { t.FailNow() }
}

