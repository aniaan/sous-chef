function PLUGIN:BackendListVersions(ctx)
  local cmd = require("cmd")
  local sc = require("lib")
  local tool = ctx.tool

  local bin = sc.get_binary()

  local command = string.format("%s list-versions --tool %s", bin, tool)
  local stdout = cmd.exec(command)

  local versions = {}
  for line in stdout:gmatch("[^\r\n]+") do
    table.insert(versions, line)
  end

  return { versions = versions }
end
