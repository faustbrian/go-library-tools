require "json"
require "digest"
require "fileutils"
require "open3"
require "time"

task = File.realpath(File.dirname(__FILE__))
abort "wrong task" unless task == "/private/tmp/identity-capability-focused.uWjCtu"
package = "/Users/brian/Developer/.worktrees/golib-go-capability-identity-platform"
control = "/Users/brian/Developer/.worktrees/golib-go-library-tools-identity-platform"
env = {"GOCACHE" => task + "/build", "GOMODCACHE" => task + "/mod", "GOTMPDIR" => task + "/tmp", "GOBIN" => task + "/bin", "GOTOOLCHAIN" => "go1.26.6", "GOWORK" => "off", "GOFLAGS" => "-p=2", "GOMAXPROCS" => "2"}
%w[build mod tmp bin db].each { |name| FileUtils.mkdir_p(task + "/" + name) }
containers = []
results = []
def capture(env, cwd, *argv)
  out, err, status = Open3.capture3(env, *argv, chdir: cwd)
  abort "acquisition failed: #{argv.first}" unless status.success?
  out
end
def execute(env, cwd, task, results, name, argv)
  started = Time.now.utc.iso8601
  status = nil
  File.open(task + "/" + name + ".log", "wb") do |log|
    Open3.popen2e(env, *argv, chdir: cwd) do |input, stream, waiter|
      input.close
      stream.each_line { |line| log.write(line); log.flush; STDOUT.write(line); STDOUT.flush }
      status = waiter.value
    end
  end
  results << {name: name, argv: argv, started_at: started, completed_at: Time.now.utc.iso8601, exit_code: status.exitstatus, log_sha256: Digest::SHA256.file(task + "/" + name + ".log").hexdigest}
  File.binwrite(task + "/results.json", JSON.pretty_generate(results) + "\n")
  abort "focused gate failed: #{name}" unless status.success?
end
begin
  revision = capture(env, package, "git", "rev-parse", "HEAD").strip
  abort "unexpected revision" unless revision == "c43af8abe83f33119ce2b13ec978e556408344f3"
  abort "dirty package" unless capture(env, package, "git", "status", "--porcelain").empty?
  zip = control + "/.ai/identity-platform/evidence/password-g0-narrow/vulndb.zip"
  abort "database changed" unless Digest::SHA256.file(zip).hexdigest == "faa8c45538d259ae1fd3c34ae20632ee738d1190836c84ac95d5f2bc604d85bc"
  capture(env, task, "/usr/bin/unzip", "-q", zip, "-d", task + "/db")
  capture(env, package, "/opt/homebrew/bin/go", "install", "golang.org/x/vuln/cmd/govulncheck@v1.6.0")
  images = %w[postgres:18.4-alpine valkey/valkey:9.1.0-alpine].map do |tag|
    row = JSON.parse(capture(env, package, "/usr/local/bin/docker", "image", "inspect", tag)).first
    {tag: tag, id: row.fetch("Id"), repo_digests: row.fetch("RepoDigests"), architecture: row.fetch("Architecture"), os: row.fetch("Os")}
  end
  inputs = {revision: revision, package_tree: capture(env, package, "git", "rev-parse", "HEAD^{tree}").strip, environment: JSON.parse(capture(env, package, "/opt/homebrew/bin/go", "env", "-json", "GOVERSION", "GOOS", "GOARCH", "CGO_ENABLED", "GOFLAGS", "GOWORK")), database: {source: "https://vuln.go.dev/vulndb.zip", custody: "password-g0-narrow/vulndb.zip", sha256: Digest::SHA256.file(zip).hexdigest}, scanner: {sha256: Digest::SHA256.file(task + "/bin/govulncheck").hexdigest, build_info: capture(env, package, "/opt/homebrew/bin/go", "version", "-m", task + "/bin/govulncheck")}, images: images, clean_consumer_script_sha256: Digest::SHA256.file(package + "/scripts/check-clean-consumer.sh").hexdigest, native_input_manifest_sha256: Digest::SHA256.file(control + "/.ai/identity-platform/evidence/capability-g0-coordinator-native/input-manifest.json").hexdigest}
  File.binwrite(task + "/inputs.json", JSON.pretty_generate(inputs) + "\n")
  execute(env, package, task, results, "clean-consumer", ["/usr/bin/env", "bash", "scripts/check-clean-consumer.sh"])
  execute(env, package, task, results, "vulnerability", [task + "/bin/govulncheck", "-db", "file://" + task + "/db", "-json", "./..."])
  images.each_with_index do |image, index|
    name = "identity-capability-focused-uWjCtu-#{index}".downcase
    argv = ["/usr/local/bin/docker", "run", "--detach", "--name", name]
    if index == 0
      argv += ["-p", "127.0.0.1::5432", "-e", "POSTGRES_USER=golib", "-e", "POSTGRES_PASSWORD=golib", "-e", "POSTGRES_DB=golib"]
    else
      argv += ["-p", "127.0.0.1::6379"]
    end
    containers << name
    capture(env, package, *(argv + [image.fetch(:id)]))
    abort "fixture identity mismatch" unless capture(env, package, "/usr/local/bin/docker", "inspect", "--format", "{{.Image}}", name).strip == image.fetch(:id)
    ready = false
    45.times do
      probe = index == 0 ? ["pg_isready", "-U", "golib"] : ["valkey-cli", "ping"]
      _, _, status = Open3.capture3(env, "/usr/local/bin/docker", "exec", name, *probe)
      if status.success?
        ready = true
        break
      end
      sleep 0.5
    end
    abort "fixture readiness failed" unless ready
    port = capture(env, package, "/usr/local/bin/docker", "port", name, index == 0 ? "5432/tcp" : "6379/tcp").strip.split(":").last
    if index == 0
      env["POSTGRES_URL"] = "postgres://golib:golib@127.0.0.1:#{port}/golib?sslmode=disable"
    else
      env["VALKEY_ADDR"] = "127.0.0.1:#{port}"
    end
  end
  execute(env, package, task, results, "pinned-services", ["/opt/homebrew/bin/go", "test", "-count=1", "-race", "-tags=integration", "./postgres", "./valkey"])
ensure
  containers.reverse_each do |name|
    _, _, status = Open3.capture3("/usr/local/bin/docker", "rm", "--force", "--volumes", name)
    warn "fixture cleanup failed: #{name}" unless status.success?
  end
  %w[build mod tmp bin db].each do |name|
    path = task + "/" + name
    FileUtils.chmod_R("u+w", path) if File.exist?(path)
    FileUtils.remove_entry(path) if File.exist?(path)
  end
end
