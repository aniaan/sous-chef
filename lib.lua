local M = {}
local cmd = require("cmd")
local file = require("file")

function M.get_binary()
  local version = PLUGIN.version

  -- RUNTIME is a global table injected by the runtime (mise/vfox)
  local plugin_dir = RUNTIME.pluginDirPath
  local os_name = RUNTIME.osType:lower()
  local arch_name = RUNTIME.archType:lower()

  local bin_dir = file.join_path(plugin_dir, "bin")
  local bin_name = "sous-chef-v" .. version
  local bin_path = file.join_path(bin_dir, bin_name)

  if file.exists(bin_path) then
    return bin_path
  end

  -- Check if development binary exists (for local testing)
  local dev_bin = file.join_path(plugin_dir, "sous-chef")
  if file.exists(dev_bin) then
    return dev_bin
  end

  -- Construct download URL
  local url = string.format(
    "https://github.com/aniaan/sous-chef/releases/download/v%s/sous-chef-%s-%s",
    version,
    os_name,
    arch_name
  )

  print("Bootstrapping sous-chef backend (v" .. version .. ")...")

  -- Create bin dir if it doesn't exist
  cmd.exec("mkdir -p " .. bin_dir)

  -- Clean up old versions to save space
  pcall(function()
    local stdout = cmd.exec("ls " .. bin_dir)
    for f in stdout:gmatch("[^\r\n]+") do
      if f:match("^sous%-chef%-v") and f ~= bin_name then
        file.remove(file.join_path(bin_dir, f))
      end
    end
  end)

  -- Download
  print("Downloading " .. url)
  cmd.exec("curl -fLo " .. bin_path .. " " .. url)

  -- Make executable
  cmd.exec("chmod +x " .. bin_path)

  return bin_path
end

return M
