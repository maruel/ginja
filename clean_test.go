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


const char kTestFilename[] = "CleanTest-tempfile"

type CleanTest struct {
  VirtualFileSystem fs_
  BuildConfig config_
  func (c *CleanTest) SetUp() {
    config_.verbosity = BuildConfig::QUIET
  }
}

func TestCleanTest_CleanAll(t *testing.T) {
  ASSERT_NO_FATAL_FAILURE(AssertParse(&state_, "build in1: cat src1\n" "build out1: cat in1\n" "build in2: cat src2\n" "build out2: cat in2\n"))
  fs_.Create("in1", "")
  fs_.Create("out1", "")
  fs_.Create("in2", "")
  fs_.Create("out2", "")

  ASSERT_EQ(0, cleaner.cleaned_files_count())
  EXPECT_EQ(0, cleaner.CleanAll())
  EXPECT_EQ(4, cleaner.cleaned_files_count())
  EXPECT_EQ(4u, fs_.files_removed_.size())

  // Check they are removed.
  string err
  EXPECT_EQ(0, fs_.Stat("in1", &err))
  EXPECT_EQ(0, fs_.Stat("out1", &err))
  EXPECT_EQ(0, fs_.Stat("in2", &err))
  EXPECT_EQ(0, fs_.Stat("out2", &err))
  fs_.files_removed_ = nil

  EXPECT_EQ(0, cleaner.CleanAll())
  EXPECT_EQ(0, cleaner.cleaned_files_count())
  EXPECT_EQ(0u, fs_.files_removed_.size())
}

func TestCleanTest_CleanAllDryRun(t *testing.T) {
  ASSERT_NO_FATAL_FAILURE(AssertParse(&state_, "build in1: cat src1\n" "build out1: cat in1\n" "build in2: cat src2\n" "build out2: cat in2\n"))
  fs_.Create("in1", "")
  fs_.Create("out1", "")
  fs_.Create("in2", "")
  fs_.Create("out2", "")

  config_.dry_run = true

  ASSERT_EQ(0, cleaner.cleaned_files_count())
  EXPECT_EQ(0, cleaner.CleanAll())
  EXPECT_EQ(4, cleaner.cleaned_files_count())
  EXPECT_EQ(0u, fs_.files_removed_.size())

  // Check they are not removed.
  string err
  EXPECT_LT(0, fs_.Stat("in1", &err))
  EXPECT_LT(0, fs_.Stat("out1", &err))
  EXPECT_LT(0, fs_.Stat("in2", &err))
  EXPECT_LT(0, fs_.Stat("out2", &err))
  fs_.files_removed_ = nil

  EXPECT_EQ(0, cleaner.CleanAll())
  EXPECT_EQ(4, cleaner.cleaned_files_count())
  EXPECT_EQ(0u, fs_.files_removed_.size())
}

func TestCleanTest_CleanTarget(t *testing.T) {
  ASSERT_NO_FATAL_FAILURE(AssertParse(&state_, "build in1: cat src1\n" "build out1: cat in1\n" "build in2: cat src2\n" "build out2: cat in2\n"))
  fs_.Create("in1", "")
  fs_.Create("out1", "")
  fs_.Create("in2", "")
  fs_.Create("out2", "")

  ASSERT_EQ(0, cleaner.cleaned_files_count())
  ASSERT_EQ(0, cleaner.CleanTarget("out1"))
  EXPECT_EQ(2, cleaner.cleaned_files_count())
  EXPECT_EQ(2u, fs_.files_removed_.size())

  // Check they are removed.
  string err
  EXPECT_EQ(0, fs_.Stat("in1", &err))
  EXPECT_EQ(0, fs_.Stat("out1", &err))
  EXPECT_LT(0, fs_.Stat("in2", &err))
  EXPECT_LT(0, fs_.Stat("out2", &err))
  fs_.files_removed_ = nil

  ASSERT_EQ(0, cleaner.CleanTarget("out1"))
  EXPECT_EQ(0, cleaner.cleaned_files_count())
  EXPECT_EQ(0u, fs_.files_removed_.size())
}

func TestCleanTest_CleanTargetDryRun(t *testing.T) {
  ASSERT_NO_FATAL_FAILURE(AssertParse(&state_, "build in1: cat src1\n" "build out1: cat in1\n" "build in2: cat src2\n" "build out2: cat in2\n"))
  fs_.Create("in1", "")
  fs_.Create("out1", "")
  fs_.Create("in2", "")
  fs_.Create("out2", "")

  config_.dry_run = true

  ASSERT_EQ(0, cleaner.cleaned_files_count())
  ASSERT_EQ(0, cleaner.CleanTarget("out1"))
  EXPECT_EQ(2, cleaner.cleaned_files_count())
  EXPECT_EQ(0u, fs_.files_removed_.size())

  // Check they are not removed.
  string err
  EXPECT_LT(0, fs_.Stat("in1", &err))
  EXPECT_LT(0, fs_.Stat("out1", &err))
  EXPECT_LT(0, fs_.Stat("in2", &err))
  EXPECT_LT(0, fs_.Stat("out2", &err))
  fs_.files_removed_ = nil

  ASSERT_EQ(0, cleaner.CleanTarget("out1"))
  EXPECT_EQ(2, cleaner.cleaned_files_count())
  EXPECT_EQ(0u, fs_.files_removed_.size())
}

func TestCleanTest_CleanRule(t *testing.T) {
  ASSERT_NO_FATAL_FAILURE(AssertParse(&state_, "rule cat_e\n" "  command = cat -e $in > $out\n" "build in1: cat_e src1\n" "build out1: cat in1\n" "build in2: cat_e src2\n" "build out2: cat in2\n"))
  fs_.Create("in1", "")
  fs_.Create("out1", "")
  fs_.Create("in2", "")
  fs_.Create("out2", "")

  ASSERT_EQ(0, cleaner.cleaned_files_count())
  ASSERT_EQ(0, cleaner.CleanRule("cat_e"))
  EXPECT_EQ(2, cleaner.cleaned_files_count())
  EXPECT_EQ(2u, fs_.files_removed_.size())

  // Check they are removed.
  string err
  EXPECT_EQ(0, fs_.Stat("in1", &err))
  EXPECT_LT(0, fs_.Stat("out1", &err))
  EXPECT_EQ(0, fs_.Stat("in2", &err))
  EXPECT_LT(0, fs_.Stat("out2", &err))
  fs_.files_removed_ = nil

  ASSERT_EQ(0, cleaner.CleanRule("cat_e"))
  EXPECT_EQ(0, cleaner.cleaned_files_count())
  EXPECT_EQ(0u, fs_.files_removed_.size())
}

func TestCleanTest_CleanRuleDryRun(t *testing.T) {
  ASSERT_NO_FATAL_FAILURE(AssertParse(&state_, "rule cat_e\n" "  command = cat -e $in > $out\n" "build in1: cat_e src1\n" "build out1: cat in1\n" "build in2: cat_e src2\n" "build out2: cat in2\n"))
  fs_.Create("in1", "")
  fs_.Create("out1", "")
  fs_.Create("in2", "")
  fs_.Create("out2", "")

  config_.dry_run = true

  ASSERT_EQ(0, cleaner.cleaned_files_count())
  ASSERT_EQ(0, cleaner.CleanRule("cat_e"))
  EXPECT_EQ(2, cleaner.cleaned_files_count())
  EXPECT_EQ(0u, fs_.files_removed_.size())

  // Check they are not removed.
  string err
  EXPECT_LT(0, fs_.Stat("in1", &err))
  EXPECT_LT(0, fs_.Stat("out1", &err))
  EXPECT_LT(0, fs_.Stat("in2", &err))
  EXPECT_LT(0, fs_.Stat("out2", &err))
  fs_.files_removed_ = nil

  ASSERT_EQ(0, cleaner.CleanRule("cat_e"))
  EXPECT_EQ(2, cleaner.cleaned_files_count())
  EXPECT_EQ(0u, fs_.files_removed_.size())
}

func TestCleanTest_CleanRuleGenerator(t *testing.T) {
  ASSERT_NO_FATAL_FAILURE(AssertParse(&state_, "rule regen\n" "  command = cat $in > $out\n" "  generator = 1\n" "build out1: cat in1\n" "build out2: regen in2\n"))
  fs_.Create("out1", "")
  fs_.Create("out2", "")

  EXPECT_EQ(0, cleaner.CleanAll())
  EXPECT_EQ(1, cleaner.cleaned_files_count())
  EXPECT_EQ(1u, fs_.files_removed_.size())

  fs_.Create("out1", "")

  EXPECT_EQ(0, cleaner.CleanAll(/*generator=*/true))
  EXPECT_EQ(2, cleaner.cleaned_files_count())
  EXPECT_EQ(2u, fs_.files_removed_.size())
}

func TestCleanTest_CleanDepFile(t *testing.T) {
  ASSERT_NO_FATAL_FAILURE(AssertParse(&state_, "rule cc\n" "  command = cc $in > $out\n" "  depfile = $out.d\n" "build out1: cc in1\n"))
  fs_.Create("out1", "")
  fs_.Create("out1.d", "")

  EXPECT_EQ(0, cleaner.CleanAll())
  EXPECT_EQ(2, cleaner.cleaned_files_count())
  EXPECT_EQ(2u, fs_.files_removed_.size())
}

func TestCleanTest_CleanDepFileOnCleanTarget(t *testing.T) {
  ASSERT_NO_FATAL_FAILURE(AssertParse(&state_, "rule cc\n" "  command = cc $in > $out\n" "  depfile = $out.d\n" "build out1: cc in1\n"))
  fs_.Create("out1", "")
  fs_.Create("out1.d", "")

  EXPECT_EQ(0, cleaner.CleanTarget("out1"))
  EXPECT_EQ(2, cleaner.cleaned_files_count())
  EXPECT_EQ(2u, fs_.files_removed_.size())
}

func TestCleanTest_CleanDepFileOnCleanRule(t *testing.T) {
  ASSERT_NO_FATAL_FAILURE(AssertParse(&state_, "rule cc\n" "  command = cc $in > $out\n" "  depfile = $out.d\n" "build out1: cc in1\n"))
  fs_.Create("out1", "")
  fs_.Create("out1.d", "")

  EXPECT_EQ(0, cleaner.CleanRule("cc"))
  EXPECT_EQ(2, cleaner.cleaned_files_count())
  EXPECT_EQ(2u, fs_.files_removed_.size())
}

func TestCleanTest_CleanDyndep(t *testing.T) {
  // Verify that a dyndep file can be loaded to discover a new output
  // to be cleaned.
  ASSERT_NO_FATAL_FAILURE(AssertParse(&state_, "build out: cat in || dd\n" "  dyndep = dd\n" ))
  fs_.Create("in", "")
  fs_.Create("dd", "ninja_dyndep_version = 1\n" "build out | out.imp: dyndep\n" )
  fs_.Create("out", "")
  fs_.Create("out.imp", "")

  ASSERT_EQ(0, cleaner.cleaned_files_count())
  EXPECT_EQ(0, cleaner.CleanAll())
  EXPECT_EQ(2, cleaner.cleaned_files_count())
  EXPECT_EQ(2u, fs_.files_removed_.size())

  string err
  EXPECT_EQ(0, fs_.Stat("out", &err))
  EXPECT_EQ(0, fs_.Stat("out.imp", &err))
}

func TestCleanTest_CleanDyndepMissing(t *testing.T) {
  // Verify that a missing dyndep file is tolerated.
  ASSERT_NO_FATAL_FAILURE(AssertParse(&state_, "build out: cat in || dd\n" "  dyndep = dd\n" ))
  fs_.Create("in", "")
  fs_.Create("out", "")
  fs_.Create("out.imp", "")

  ASSERT_EQ(0, cleaner.cleaned_files_count())
  EXPECT_EQ(0, cleaner.CleanAll())
  EXPECT_EQ(1, cleaner.cleaned_files_count())
  EXPECT_EQ(1u, fs_.files_removed_.size())

  string err
  EXPECT_EQ(0, fs_.Stat("out", &err))
  EXPECT_EQ(1, fs_.Stat("out.imp", &err))
}

func TestCleanTest_CleanRspFile(t *testing.T) {
  ASSERT_NO_FATAL_FAILURE(AssertParse(&state_, "rule cc\n" "  command = cc $in > $out\n" "  rspfile = $rspfile\n" "  rspfile_content=$in\n" "build out1: cc in1\n" "  rspfile = cc1.rsp\n"))
  fs_.Create("out1", "")
  fs_.Create("cc1.rsp", "")

  EXPECT_EQ(0, cleaner.CleanAll())
  EXPECT_EQ(2, cleaner.cleaned_files_count())
  EXPECT_EQ(2u, fs_.files_removed_.size())
}

func TestCleanTest_CleanRsp(t *testing.T) {
  ASSERT_NO_FATAL_FAILURE(AssertParse(&state_, "rule cat_rsp \n" "  command = cat $rspfile > $out\n" "  rspfile = $rspfile\n" "  rspfile_content = $in\n" "build in1: cat src1\n" "build out1: cat in1\n" "build in2: cat_rsp src2\n" "  rspfile=in2.rsp\n" "build out2: cat_rsp in2\n" "  rspfile=out2.rsp\n" ))
  fs_.Create("in1", "")
  fs_.Create("out1", "")
  fs_.Create("in2.rsp", "")
  fs_.Create("out2.rsp", "")
  fs_.Create("in2", "")
  fs_.Create("out2", "")

  ASSERT_EQ(0, cleaner.cleaned_files_count())
  ASSERT_EQ(0, cleaner.CleanTarget("out1"))
  EXPECT_EQ(2, cleaner.cleaned_files_count())
  ASSERT_EQ(0, cleaner.CleanTarget("in2"))
  EXPECT_EQ(2, cleaner.cleaned_files_count())
  ASSERT_EQ(0, cleaner.CleanRule("cat_rsp"))
  EXPECT_EQ(2, cleaner.cleaned_files_count())

  EXPECT_EQ(6u, fs_.files_removed_.size())

  // Check they are removed.
  string err
  EXPECT_EQ(0, fs_.Stat("in1", &err))
  EXPECT_EQ(0, fs_.Stat("out1", &err))
  EXPECT_EQ(0, fs_.Stat("in2", &err))
  EXPECT_EQ(0, fs_.Stat("out2", &err))
  EXPECT_EQ(0, fs_.Stat("in2.rsp", &err))
  EXPECT_EQ(0, fs_.Stat("out2.rsp", &err))
}

func TestCleanTest_CleanFailure(t *testing.T) {
  ASSERT_NO_FATAL_FAILURE(AssertParse(&state_, "build dir: cat src1\n"))
  fs_.MakeDir("dir")
  EXPECT_NE(0, cleaner.CleanAll())
}

func TestCleanTest_CleanPhony(t *testing.T) {
  string err
  ASSERT_NO_FATAL_FAILURE(AssertParse(&state_, "build phony: phony t1 t2\n" "build t1: cat\n" "build t2: cat\n"))

  fs_.Create("phony", "")
  fs_.Create("t1", "")
  fs_.Create("t2", "")

  EXPECT_EQ(0, cleaner.CleanAll())
  EXPECT_EQ(2, cleaner.cleaned_files_count())
  EXPECT_LT(0, fs_.Stat("phony", &err))

  fs_.Create("t1", "")
  fs_.Create("t2", "")

  // Check that CleanTarget does not remove "phony".
  EXPECT_EQ(0, cleaner.CleanTarget("phony"))
  EXPECT_EQ(2, cleaner.cleaned_files_count())
  EXPECT_LT(0, fs_.Stat("phony", &err))
}

func TestCleanTest_CleanDepFileAndRspFileWithSpaces(t *testing.T) {
  ASSERT_NO_FATAL_FAILURE(AssertParse(&state_, "rule cc_dep\n" "  command = cc $in > $out\n" "  depfile = $out.d\n" "rule cc_rsp\n" "  command = cc $in > $out\n" "  rspfile = $out.rsp\n" "  rspfile_content = $in\n" "build out$ 1: cc_dep in$ 1\n" "build out$ 2: cc_rsp in$ 1\n" ))
  fs_.Create("out 1", "")
  fs_.Create("out 2", "")
  fs_.Create("out 1.d", "")
  fs_.Create("out 2.rsp", "")

  EXPECT_EQ(0, cleaner.CleanAll())
  EXPECT_EQ(4, cleaner.cleaned_files_count())
  EXPECT_EQ(4u, fs_.files_removed_.size())

  string err
  EXPECT_EQ(0, fs_.Stat("out 1", &err))
  EXPECT_EQ(0, fs_.Stat("out 2", &err))
  EXPECT_EQ(0, fs_.Stat("out 1.d", &err))
  EXPECT_EQ(0, fs_.Stat("out 2.rsp", &err))
}

type CleanDeadTest struct {
  func (c *CleanDeadTest) SetUp() {
    // In case a crashing test left a stale file behind.
    unlink(kTestFilename)
    CleanTest::SetUp()
  }
  func (c *CleanDeadTest) TearDown() {
    unlink(kTestFilename)
  }
  virtual bool IsPathDead(StringPiece) const { return false; }
}

func TestCleanDeadTest_CleanDead(t *testing.T) {
  State state
  ASSERT_NO_FATAL_FAILURE(AssertParse(&state, "rule cat\n" "  command = cat $in > $out\n" "build out1: cat in\n" "build out2: cat in\n" ))
  ASSERT_NO_FATAL_FAILURE(AssertParse(&state_, "build out2: cat in\n" ))
  fs_.Create("in", "")
  fs_.Create("out1", "")
  fs_.Create("out2", "")

  BuildLog log1
  string err
  EXPECT_TRUE(log1.OpenForWrite(kTestFilename, *this, &err))
  ASSERT_EQ("", err)
  log1.RecordCommand(state.edges_[0], 15, 18)
  log1.RecordCommand(state.edges_[1], 20, 25)
  log1.Close()

  BuildLog log2
  EXPECT_TRUE(log2.Load(kTestFilename, &err))
  ASSERT_EQ("", err)
  ASSERT_EQ(2u, log2.entries().size())
  ASSERT_TRUE(log2.LookupByOutput("out1"))
  ASSERT_TRUE(log2.LookupByOutput("out2"))

  EXPECT_EQ(0, cleaner1.CleanDead(log2.entries()))
  EXPECT_EQ(0, cleaner1.cleaned_files_count())
  EXPECT_EQ(0u, fs_.files_removed_.size())
  EXPECT_NE(0, fs_.Stat("in", &err))
  EXPECT_NE(0, fs_.Stat("out1", &err))
  EXPECT_NE(0, fs_.Stat("out2", &err))

  EXPECT_EQ(0, cleaner2.CleanDead(log2.entries()))
  EXPECT_EQ(1, cleaner2.cleaned_files_count())
  EXPECT_EQ(1u, fs_.files_removed_.size())
  EXPECT_EQ("out1", *(fs_.files_removed_.begin()))
  EXPECT_NE(0, fs_.Stat("in", &err))
  EXPECT_EQ(0, fs_.Stat("out1", &err))
  EXPECT_NE(0, fs_.Stat("out2", &err))

  // Nothing to do now.
  EXPECT_EQ(0, cleaner2.CleanDead(log2.entries()))
  EXPECT_EQ(0, cleaner2.cleaned_files_count())
  EXPECT_EQ(1u, fs_.files_removed_.size())
  EXPECT_EQ("out1", *(fs_.files_removed_.begin()))
  EXPECT_NE(0, fs_.Stat("in", &err))
  EXPECT_EQ(0, fs_.Stat("out1", &err))
  EXPECT_NE(0, fs_.Stat("out2", &err))
  log2.Close()
}

func TestCleanDeadTest_CleanDeadPreservesInputs(t *testing.T) {
  State state
  ASSERT_NO_FATAL_FAILURE(AssertParse(&state, "rule cat\n" "  command = cat $in > $out\n" "build out1: cat in\n" "build out2: cat in\n" ))
  // This manifest does not build out1 anymore, but makes
  // it an implicit input. CleanDead should detect this
  // and preserve it.
  ASSERT_NO_FATAL_FAILURE(AssertParse(&state_, "build out2: cat in | out1\n" ))
  fs_.Create("in", "")
  fs_.Create("out1", "")
  fs_.Create("out2", "")

  BuildLog log1
  string err
  EXPECT_TRUE(log1.OpenForWrite(kTestFilename, *this, &err))
  ASSERT_EQ("", err)
  log1.RecordCommand(state.edges_[0], 15, 18)
  log1.RecordCommand(state.edges_[1], 20, 25)
  log1.Close()

  BuildLog log2
  EXPECT_TRUE(log2.Load(kTestFilename, &err))
  ASSERT_EQ("", err)
  ASSERT_EQ(2u, log2.entries().size())
  ASSERT_TRUE(log2.LookupByOutput("out1"))
  ASSERT_TRUE(log2.LookupByOutput("out2"))

  EXPECT_EQ(0, cleaner1.CleanDead(log2.entries()))
  EXPECT_EQ(0, cleaner1.cleaned_files_count())
  EXPECT_EQ(0u, fs_.files_removed_.size())
  EXPECT_NE(0, fs_.Stat("in", &err))
  EXPECT_NE(0, fs_.Stat("out1", &err))
  EXPECT_NE(0, fs_.Stat("out2", &err))

  EXPECT_EQ(0, cleaner2.CleanDead(log2.entries()))
  EXPECT_EQ(0, cleaner2.cleaned_files_count())
  EXPECT_EQ(0u, fs_.files_removed_.size())
  EXPECT_NE(0, fs_.Stat("in", &err))
  EXPECT_NE(0, fs_.Stat("out1", &err))
  EXPECT_NE(0, fs_.Stat("out2", &err))

  // Nothing to do now.
  EXPECT_EQ(0, cleaner2.CleanDead(log2.entries()))
  EXPECT_EQ(0, cleaner2.cleaned_files_count())
  EXPECT_EQ(0u, fs_.files_removed_.size())
  EXPECT_NE(0, fs_.Stat("in", &err))
  EXPECT_NE(0, fs_.Stat("out1", &err))
  EXPECT_NE(0, fs_.Stat("out2", &err))
  log2.Close()
}

