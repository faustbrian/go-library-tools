require "digest"
require "json"
require "fileutils"
require "open3"
require "time"
require "/Users/brian/Developer/.worktrees/golib-go-library-tools-identity-platform/.ai/identity-platform/repository_input_manifest"

task = File.realpath(File.dirname(__FILE__))
abort "wrong task root" unless task == "/private/tmp/identity-capability-coordinator-gate.tZayOh"
control = "/Users/brian/Developer/.worktrees/golib-go-library-tools-identity-platform"
package = "/Users/brian/Developer/.worktrees/golib-go-capability-identity-platform"
unit = "primitive/capability-identity-contracts"
env = {"GOCACHE" => task + "/build", "GOMODCACHE" => task + "/mod", "GOTMPDIR" => task + "/tmp", "GOTOOLCHAIN" => "go1.26.6", "GOWORK" => "off", "GOFLAGS" => "-p=2", "GOMAXPROCS" => "2"}
%w[build mod tmp].each { |name| FileUtils.mkdir_p(task + "/" + name) }
def capture(env, cwd, *args)
  out, err, status = Open3.capture3(env, *args, chdir: cwd)
  abort "input acquisition failed: #{args.first}" unless status.success?
  out
end
begin
  revision = capture(env, package, "git", "rev-parse", "HEAD").strip
  abort "changed package checkpoint" unless revision == "c43af8abe83f33119ce2b13ec978e556408344f3"
  abort "dirty package" unless capture(env, package, "git", "status", "--porcelain").empty?
  control_revision = capture(env, control, "git", "rev-parse", "HEAD").strip
  contract = JSON.parse(File.read(control + "/.ai/identity-platform/fragments/public_contracts_meta.json")).fetch("external_contracts").find { |row| row["owner"] == "capability" }
  applicability = JSON.parse(File.read(control + "/.ai/identity-platform/VERIFICATION_APPLICABILITY.json"))
  observations = {
    "package-environment" => JSON.parse(capture(env, package, "/opt/homebrew/bin/go", "env", "-json", "GOVERSION", "GOOS", "GOARCH", "CGO_ENABLED", "GOFLAGS", "GOWORK")),
    "python-interop" => {"path" => "/usr/bin/python3", "sha256" => Digest::SHA256.file("/usr/bin/python3").hexdigest, "version" => capture(env, package, "/usr/bin/python3", "--version")},
    "native-launcher" => {"path" => "/usr/bin/make", "sha256" => Digest::SHA256.file("/usr/bin/make").hexdigest, "version" => capture(env, package, "/usr/bin/make", "--version"), "argv" => ["/usr/bin/make", "ci"]},
    "selected-primitive-contract" => contract,
    "selected-applicability" => applicability,
    "host" => {"os" => capture(env, package, "/usr/bin/sw_vers"), "gomaxprocs" => "2"}
  }
  roots = %w[COMMON_REQUIREMENTS.md ORCHESTRATOR_GOAL.md goals/primitive-capability-identity-contracts.md public_contract_repository.rb repository_input_manifest.rb repository_topology.rb].map { |p| {path: ".ai/identity-platform/" + p, owner: "coordinator", reason: "Selected unit requirements and existing canonical input readers"} }
  manifest = IdentityPlatformRepositoryInputManifest.build(repositories: [
    {repository_identity: "github.com/faustbrian/go-library-tools", repository_root: control, revision: control_revision, input_roots: roots},
    {repository_identity: "github.com/faustbrian/go-capability", repository_root: package, revision: revision, input_roots: [{path: ".", owner: unit, reason: "Complete committed package sources, fixtures, dependencies, native gate scripts and configuration"}]}
  ], environment_inputs: observations.map { |name, value| {environment_id: "environment:" + name, content_identity: "sha256:" + Digest::SHA256.hexdigest(JSON.generate(value)), owner: "coordinator", reason: "Directly acquired before coordinator native module execution; external live vulnerability feed not frozen by this record"} })
  File.binwrite(task + "/input-manifest.json", JSON.pretty_generate(manifest) + "\n")
  File.binwrite(task + "/environment-inputs.json", JSON.pretty_generate(observations) + "\n")
  root = IdentityPlatformRepositoryInputManifest.fingerprint(manifest)
  puts "Coordinator capability revision #{revision}; pre-execution input root #{root}"
  started = Time.now.utc.iso8601
  status = nil
  File.open(task + "/native.log", "wb") do |log|
    Open3.popen2e(env, "/usr/bin/make", "ci", chdir: package) do |stdin, stream, waiter|
      stdin.close
      stream.each_line { |line| log.write(line); log.flush; STDOUT.write(line); STDOUT.flush }
      status = waiter.value
    end
  end
  File.binwrite(task + "/native-result.json", JSON.pretty_generate({revision: revision, input_root: root, started_at: started, completed_at: Time.now.utc.iso8601, argv: ["/usr/bin/make", "ci"], exit_code: status.exitstatus, merged_stream_sha256: Digest::SHA256.file(task + "/native.log").hexdigest}) + "\n")
  puts "Coordinator capability native exit #{status.exitstatus}"
ensure
  %w[build mod tmp].each do |name|
    path = task + "/" + name
    FileUtils.chmod_R("u+w", path) if File.exist?(path)
    FileUtils.remove_entry(path) if File.exist?(path)
  end
end
