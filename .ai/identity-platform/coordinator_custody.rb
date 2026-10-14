# frozen_string_literal: true

require "open3"
require_relative "repository_input_manifest"

module IdentityPlatformCoordinatorCustody
  class InvalidCustody < StandardError; end

  module_function

  def resolve_commit(repository_root:, repository_identity:, revision:, input_roots:)
    document = IdentityPlatformRepositoryInputManifest.build(
      repositories: [{
        repository_root: repository_root,
        repository_identity: repository_identity,
        revision: revision,
        input_roots: input_roots
      }],
      environment_inputs: []
    )
    document.fetch("repositories").fetch(0)
  rescue IdentityPlatformRepositoryInputManifest::InvalidManifest => e
    raise InvalidCustody, e.message
  end

  def ancestor?(ancestor:, descendant:, repository_root:)
    ancestor_identity = required_identity(ancestor[:repository_identity])
    descendant_identity = required_identity(descendant[:repository_identity])
    unless ancestor_identity == descendant_identity
      raise InvalidCustody,
        "cross-repository ancestry is forbidden: #{ancestor_identity} -> #{descendant_identity}"
    end

    resolve_commit(
      repository_root: repository_root,
      repository_identity: ancestor_identity,
      revision: ancestor[:revision],
      input_roots: [{path: ".", owner: "custody", reason: "ancestry endpoint"}]
    )
    resolve_commit(
      repository_root: repository_root,
      repository_identity: descendant_identity,
      revision: descendant[:revision],
      input_roots: [{path: ".", owner: "custody", reason: "ancestry endpoint"}]
    )
    _output, _error, status = Open3.capture3(
      "git", "-C", repository_root, "merge-base", "--is-ancestor",
      ancestor.fetch(:revision), descendant.fetch(:revision)
    )
    status.success?
  end

  def evidence_link_errors(control_binding:, package_binding:, control_repository_root:,
                           package_repository_root:, evidence_commit:)
    errors = []
    unless control_binding.is_a?(Hash) && package_binding.is_a?(Hash)
      return ["repository custody is absent"]
    end

    control_identity = control_binding["repository_identity"]
    package_identity = package_binding["repository_identity"]
    errors << "repository custody is not distinctly scoped" if control_identity == package_identity
    snapshots = [
      ["control", control_binding, control_repository_root],
      ["package", package_binding, package_repository_root]
    ].to_h do |label, binding, repository_root|
      begin
        snapshot = resolve_commit(
          repository_root: repository_root,
          repository_identity: binding["repository_identity"],
          revision: binding["revision"],
          input_roots: [{path: ".", owner: "custody", reason: "#{label} evidence binding"}]
        )
        errors << "#{label} repository object format drifted" unless snapshot["object_format"] == binding["object_format"]
        errors << "#{label} repository tree drifted" unless snapshot["tree"] == binding["tree"]
        [label, snapshot]
      rescue InvalidCustody => e
        errors << "#{label} repository custody is invalid: #{e.message}"
        [label, nil]
      end
    end

    if snapshots["control"]
      begin
        evidence = resolve_commit(
          repository_root: control_repository_root,
          repository_identity: control_identity,
          revision: evidence_commit,
          input_roots: [{path: ".", owner: "custody", reason: "evidence record commit"}]
        )
        unless ancestor?(
          ancestor: {repository_identity: control_identity, revision: control_binding["revision"]},
          descendant: {repository_identity: control_identity, revision: evidence["revision"]},
          repository_root: control_repository_root
        )
          errors << "control evidence commit does not descend from the control execution revision"
        end
      rescue InvalidCustody => e
        errors << "control evidence commit is invalid: #{e.message}"
      end
    end

    errors
  end

  def required_identity(value)
    identity = value.to_s
    unless identity.match?(%r{\A[a-z0-9.-]+/[A-Za-z0-9._-]+(?:/[A-Za-z0-9._-]+)+\z})
      raise InvalidCustody, "repository identity #{value.inspect} is missing or invalid"
    end

    identity
  end
  private_class_method :required_identity
end
