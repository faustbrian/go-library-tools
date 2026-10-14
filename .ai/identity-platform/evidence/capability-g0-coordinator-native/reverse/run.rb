require "json"
require "digest"
require "fileutils"
require "open3"
require "time"
task = File.realpath(File.dirname(__FILE__))
abort "wrong task" unless task == "/private/tmp/identity-capability-reverse.mE8PNO"
source = "/Users/brian/Developer/golib/go-service"
capability = "/Users/brian/Developer/.worktrees/golib-go-capability-identity-platform"
env = {"GOCACHE" => task + "/build", "GOMODCACHE" => task + "/mod", "GOTMPDIR" => task + "/tmp", "GOTOOLCHAIN" => "go1.26.6", "GOWORK" => "off", "GOFLAGS" => "-p=2", "GOMAXPROCS" => "2"}
%w[build mod tmp consumer].each { |name| FileUtils.mkdir_p(task + "/" + name) }
def capture(env, cwd, *argv)
  out, err, status = Open3.capture3(env, *argv, chdir: cwd)
  abort "input acquisition failed: #{argv.first}" unless status.success?
  out
end
begin
  revision = capture(env, source, "git", "rev-parse", "HEAD").strip
  abort "unexpected consumer" unless revision == "b0a6f412d2f5aa3c3fb067222f96dc4242b8034f"
  cap_revision = capture(env, capability, "git", "rev-parse", "HEAD").strip
  abort "unexpected dependency" unless cap_revision == "c43af8abe83f33119ce2b13ec978e556408344f3"
  archive = capture(env, source, "git", "archive", revision, "integration/reference-http")
  File.binwrite(task + "/consumer.tar", archive)
  capture(env, task, "/usr/bin/tar", "-xf", task + "/consumer.tar", "-C", task + "/consumer")
  cwd = task + "/consumer/integration/reference-http"
  capture(env, cwd, "/opt/homebrew/bin/go", "mod", "edit", "-replace=github.com/faustbrian/go-capability=" + capability)
  # Only the task-owned consumer manifest is edited; all production bytes are
  # immutable Git archive bytes, and the integrated dependency stays untouched.
  capture(env, cwd, "/opt/homebrew/bin/go", "mod", "tidy")
  inputs = {consumer_repository: "github.com/faustbrian/go-service", consumer_revision: revision, consumer_subtree: capture(env, source, "git", "rev-parse", revision + ":integration/reference-http").strip, archive_sha256: Digest::SHA256.hexdigest(archive), integrated_dependency_revision: cap_revision, integrated_dependency_tree: capture(env, capability, "git", "rev-parse", "HEAD^{tree}").strip, environment: JSON.parse(capture(env, cwd, "/opt/homebrew/bin/go", "env", "-json", "GOVERSION", "GOOS", "GOARCH", "CGO_ENABLED", "GOFLAGS", "GOWORK")), temporary_go_mod_sha256: Digest::SHA256.file(cwd + "/go.mod").hexdigest, temporary_go_sum_sha256: Digest::SHA256.file(cwd + "/go.sum").hexdigest, argv: ["/opt/homebrew/bin/go", "test", "-race", "-count=1", "-timeout=90s", "./..."]}
  File.binwrite(task + "/inputs.json", JSON.pretty_generate(inputs) + "\n")
  FileUtils.cp(cwd + "/go.mod", task + "/consumer-go.mod")
  FileUtils.cp(cwd + "/go.sum", task + "/consumer-go.sum")
  started = Time.now.utc.iso8601
  out, err, status = Open3.capture3(env, *inputs.fetch(:argv), chdir: cwd)
  File.binwrite(task + "/stdout.log", out)
  File.binwrite(task + "/stderr.log", err)
  File.binwrite(task + "/result.json", JSON.pretty_generate({started_at: started, completed_at: Time.now.utc.iso8601, exit_code: status.exitstatus, inputs_sha256: Digest::SHA256.file(task + "/inputs.json").hexdigest, stdout_sha256: Digest::SHA256.hexdigest(out), stderr_sha256: Digest::SHA256.hexdigest(err)}) + "\n")
  puts out
  warn err unless err.empty?
  abort "reverse consumer failed" unless status.success?
ensure
  %w[build mod tmp consumer].each do |name|
    path = task + "/" + name
    FileUtils.chmod_R("u+w", path) if File.exist?(path)
    FileUtils.remove_entry(path) if File.exist?(path)
  end
end
