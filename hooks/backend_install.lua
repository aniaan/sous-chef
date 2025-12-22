function PLUGIN:BackendInstall(ctx)
  local cmd = require("cmd")
  local sc = require("lib")
  local tool = ctx.tool
  local version = ctx.version
  local install_path = ctx.install_path

  local bin = sc.get_binary()

  local command = string.format("%s install --tool %s --version %s --dir %s", bin, tool, version, install_path)

  cmd.exec(command)

  return {}
end
