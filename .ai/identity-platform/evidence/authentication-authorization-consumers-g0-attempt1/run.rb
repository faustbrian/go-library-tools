require "json"
require "digest"
require "fileutils"
require "open3"
require "time"

task = File.realpath(File.dirname(__FILE__))
abort "unexpected task path" unless task == "/private/tmp/identity-platform-consumer-gates.9IxpDN"
go = "/opt/homebrew/bin/go"
env = {"GOCACHE"=>task+"/build", "GOMODCACHE"=>task+"/mod", "GOTMPDIR"=>task+"/tmp", "GOWORK"=>"off", "GOFLAGS"=>"-p=2", "GOMAXPROCS"=>"2", "GOROOT"=>nil}
%w[build mod tmp snapshots results].each { |name| FileUtils.mkdir_p(task+"/"+name) }
def acquire(env, cwd, *argv)
  out, err, status = Open3.capture3(env, *argv, chdir: cwd)
  raise "acquisition failed #{argv.first}: #{err.byteslice(0,1500)}" unless status.success?
  out
end
def atomic_json(path, value)
  File.binwrite(path+".pending", JSON.pretty_generate(value)+"\n")
  File.rename(path+".pending", path)
end
repos = {
  "authentication"=>["/Users/brian/Developer/.worktrees/golib-go-authentication-identity-platform", "a5eb0e543f53610ceaece35373986cf0a4fa5891"],
  "authorization"=>["/Users/brian/Developer/.worktrees/golib-go-authorization-identity-platform", "a74c24b3846323e0c1627c6fccbb79ce25248392"],
  "capability"=>["/Users/brian/Developer/.worktrees/golib-go-capability-identity-platform", "c43af8abe83f33119ce2b13ec978e556408344f3"],
  "middleware"=>["/Users/brian/Developer/golib/go-http-middleware", nil],
  "queue"=>["/Users/brian/Developer/golib/go-queue-control-plane", nil],
  "service"=>["/Users/brian/Developer/golib/go-service", nil]
}
units = [
  ["authorization-contracts", "authorization", "integration/contracts", ["./..."], "go1.26.6"],
  ["authentication-authotel", "authentication", "authotel", ["./..."], "go1.26.6"],
  ["authentication-jwt", "authentication", "jwt", ["./..."], "go1.26.6"],
  ["authentication-oidc", "authentication", "oidc", ["./..."], "go1.26.6"],
  ["middleware-siblings", "middleware", "integration/siblings", ["./..."], "go1.27.1"],
  ["queue-authz-server", "queue", ".", ["./authz", "./server"], "go1.27.1"],
  ["service-compatibility", "service", "compatibility", ["./..."], "go1.26.6"],
  ["service-reference-http", "service", "integration/reference-http", ["./..."], "go1.26.6"]
]
identities = {}
begin
  repos.each do |name, (source, expected)|
    revision = acquire(env, source, "git", "rev-parse", "HEAD").strip
    raise "candidate changed: #{name}" if expected && revision != expected
    archive = acquire(env, source, "git", "archive", revision)
    tar = task+"/"+name+".tar"
    File.binwrite(tar, archive)
    target = task+"/snapshots/"+name
    FileUtils.mkdir_p(target)
    acquire(env, task, "/usr/bin/tar", "-xf", tar, "-C", target)
    File.unlink(tar)
    files = Dir.glob(target+"/**/*", File::FNM_DOTMATCH).select { |p| File.file?(p) }.sort.map { |p| {"path"=>p.delete_prefix(target+"/"), "sha256"=>Digest::SHA256.file(p).hexdigest} }
    identities[name] = {"repository"=>source, "revision"=>revision, "tree"=>acquire(env, source, "git", "rev-parse", revision+"^{tree}").strip, "archive_sha256"=>Digest::SHA256.hexdigest(archive), "files"=>files}
  end
  atomic_json(task+"/results/source-inputs.json", identities)
  failures = []
  units.each do |name, repo, subdir, packages, toolchain|
    result_dir = task+"/results/"+name
    FileUtils.mkdir_p(result_dir)
    cwd = File.join(task, "snapshots", repo, subdir)
    selected = env.merge("GOTOOLCHAIN"=>toolchain)
    argv = [go, "test", "-race", "-count=1", "-timeout=180s", *packages]
    begin
      %w[authentication authorization capability].each do |dependency|
        acquire(selected, cwd, go, "mod", "edit", "-replace=github.com/faustbrian/go-#{dependency}=#{task}/snapshots/#{dependency}")
      end
      acquire(selected, cwd, go, "mod", "tidy")
      %w[go.mod go.sum].each { |path| FileUtils.cp(cwd+"/"+path, result_dir+"/"+path) if File.file?(cwd+"/"+path) }
      inputs = {"consumer"=>name, "source_repository"=>repo, "subdir"=>subdir, "source_inputs_sha256"=>Digest::SHA256.file(task+"/results/source-inputs.json").hexdigest, "candidate_revisions"=>identities.slice("authentication", "authorization", "capability").transform_values { |v| v.fetch("revision") }, "go_mod_sha256"=>Digest::SHA256.file(cwd+"/go.mod").hexdigest, "go_sum_sha256"=>File.file?(cwd+"/go.sum") ? Digest::SHA256.file(cwd+"/go.sum").hexdigest : nil, "go_executable_sha256"=>Digest::SHA256.file(go).hexdigest, "tool_version"=>acquire(selected,cwd,go,"version").strip, "environment"=>JSON.parse(acquire(selected,cwd,go,"env","-json","GOVERSION","GOOS","GOARCH","CGO_ENABLED","GOTOOLCHAIN","GOWORK","GOFLAGS","GOROOT")), "argv"=>argv}
      atomic_json(result_dir+"/inputs.json", inputs)
      started = Time.now.utc.iso8601
      out, err, status = Open3.capture3(selected, *argv, chdir: cwd)
      File.binwrite(result_dir+"/stdout.log",out)
      File.binwrite(result_dir+"/stderr.log",err)
      atomic_json(result_dir+"/result.json", {"started_at"=>started,"completed_at"=>Time.now.utc.iso8601,"exit_code"=>status.exitstatus,"inputs_sha256"=>Digest::SHA256.file(result_dir+"/inputs.json").hexdigest,"stdout_sha256"=>Digest::SHA256.hexdigest(out),"stderr_sha256"=>Digest::SHA256.hexdigest(err)})
      puts "#{name}: exit #{status.exitstatus}"
      puts out
      warn err unless err.empty?
      failures << name unless status.success?
    rescue => error
      atomic_json(result_dir+"/acquisition-failure.json", {"error"=>error.message,"at"=>Time.now.utc.iso8601})
      warn "#{name}: input acquisition failed"
      failures << name
    end
    STDOUT.flush
  end
  atomic_json(task+"/results/batch-result.json", {"units"=>units.map(&:first), "failed"=>failures})
  exit(failures.empty? ? 0 : 1)
ensure
  %w[build mod tmp snapshots].each do |name|
    path = task+"/"+name
    next unless File.exist?(path)
    FileUtils.chmod_R("u+w",path)
    FileUtils.remove_entry(path)
  end
end
