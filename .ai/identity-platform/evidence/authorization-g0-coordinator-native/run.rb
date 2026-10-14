require "digest"
require "json"
require "fileutils"
require "open3"
require "time"
require "/Users/brian/Developer/.worktrees/golib-go-library-tools-identity-platform/.ai/identity-platform/repository_input_manifest"
task = File.realpath(File.dirname(__FILE__))
abort "wrong task" unless task == "/private/tmp/identity-authorization-coordinator-gate.6eAPg1"
control = "/Users/brian/Developer/.worktrees/golib-go-library-tools-identity-platform"
package = "/Users/brian/Developer/.worktrees/golib-go-authorization-identity-platform"
env = {"GOCACHE" => task + "/build", "GOMODCACHE" => task + "/mod", "GOTMPDIR" => task + "/tmp", "GOTOOLCHAIN" => "go1.26.6", "GOWORK" => "off", "GOFLAGS" => "-p=2", "GOMAXPROCS" => "2"}
%w[build mod tmp].each { |name| FileUtils.mkdir_p(task + "/" + name) }
def capture(env, cwd, *args)
  out, err, status = Open3.capture3(env, *args, chdir: cwd)
  abort "input acquisition failed: #{args.first}" unless status.success?
  out
end
begin
  revision = capture(env, package, "git", "rev-parse", "HEAD").strip
  abort "wrong package checkpoint" unless revision == "a74c24b3846323e0c1627c6fccbb79ce25248392"
  abort "dirty package" unless capture(env, package, "git", "status", "--porcelain").empty?
  control_revision = capture(env, control, "git", "rev-parse", "HEAD").strip
  tool = task + "/bin/golib"
  argv = [tool, "check", "--module", "."]
  contract = JSON.parse(File.read(control + "/.ai/identity-platform/fragments/public_contracts_meta.json")).fetch("external_contracts").find { |row| row["owner"] == "authorization" }
  observations = {"package-environment" => JSON.parse(capture(env, package, "/opt/homebrew/bin/go", "env", "-json", "GOVERSION", "GOOS", "GOARCH", "CGO_ENABLED", "GOFLAGS", "GOWORK")), "selected-primitive-contract" => contract, "selected-applicability" => JSON.parse(File.read(control + "/.ai/identity-platform/VERIFICATION_APPLICABILITY.json")), "host" => {"os" => capture(env, package, "/usr/bin/sw_vers"), "gomaxprocs" => "2"}, "native-tool" => {"argv" => argv, "path" => tool, "sha256" => Digest::SHA256.file(tool).hexdigest, "build_info" => capture(env, package, "/opt/homebrew/bin/go", "version", "-m", tool), "official_release" => "https://github.com/faustbrian/go-library-tools/releases/tag/v1.4.0", "checksums_sha256" => Digest::SHA256.file(task + "/checksums.txt").hexdigest, "archive_sha256" => Digest::SHA256.file(task + "/golib.tar.gz").hexdigest}}
  abort "unapproved tool manifest" unless observations.fetch("native-tool").fetch("checksums_sha256") == "ba1b71d41dc9b58d5bfb411bc89ae2d45b6f3f778152714bfd1afeab0bef2f33"
  roots = %w[COMMON_REQUIREMENTS.md ORCHESTRATOR_GOAL.md goals/primitive-authorization-identity-contracts.md public_contract_repository.rb repository_input_manifest.rb repository_topology.rb].map { |p| {path: ".ai/identity-platform/" + p, owner: "coordinator", reason: "Selected unit requirements and canonical input readers"} }
  manifest = IdentityPlatformRepositoryInputManifest.build(repositories: [{repository_identity: "github.com/faustbrian/go-library-tools", repository_root: control, revision: control_revision, input_roots: roots}, {repository_identity: "github.com/faustbrian/go-authorization", repository_root: package, revision: revision, input_roots: [{path: ".", owner: "primitive/authorization-identity-contracts", reason: "Complete committed family sources, fixtures, dependencies and configured native gate inputs"}]}], environment_inputs: observations.map { |name, value| {environment_id: "environment:" + name, content_identity: "sha256:" + Digest::SHA256.hexdigest(JSON.generate(value)), owner: "coordinator", reason: "Directly captured before configured root-module gate; live vulnerability feed and external fixture/tool inputs require attributable closure"} })
  File.binwrite(task + "/input-manifest.json", JSON.pretty_generate(manifest) + "\n")
  File.binwrite(task + "/environment-inputs.json", JSON.pretty_generate(observations) + "\n")
  root = IdentityPlatformRepositoryInputManifest.fingerprint(manifest)
  puts "Authorization root native check; pre-execution input #{root}"
  started = Time.now.utc.iso8601
  status = nil
  File.open(task + "/native.log", "wb") do |log|
    Open3.popen2e(env, *argv, chdir: package) do |input, stream, waiter|
      input.close
      stream.each_line { |line| log.write(line); log.flush }
      status = waiter.value
    end
  end
  File.binwrite(task + "/native-result.json", JSON.pretty_generate({revision: revision, input_root: root, started_at: started, completed_at: Time.now.utc.iso8601, argv: argv, exit_code: status.exitstatus, merged_stream_sha256: Digest::SHA256.file(task + "/native.log").hexdigest}) + "\n")
  puts "Authorization root native exit #{status.exitstatus}"
ensure
  %w[build mod tmp bin].each do |name|
    path = task + "/" + name
    FileUtils.chmod_R("u+w", path) if File.exist?(path)
    FileUtils.remove_entry(path) if File.exist?(path)
  end
end
