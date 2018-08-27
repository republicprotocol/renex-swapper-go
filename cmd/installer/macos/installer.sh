echo "<!DOCTYPE plist PUBLIC \"-//Apple//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\">
<plist version=\"1.0\">
  <dict>
    <key>Label</key>
    <string>com.renex.swapper</string>
    <key>Program</key>
    <string>$HOME/.swapper/startup.sh</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>LaunchOnlyOnce</key>        
    <true/>
    <key>StandardOutPath</key>
    <string>$HOME/.swapper/swapper.out</string>
    <key>StandardErrorPath</key>
    <string>$HOME/.swapper/swapper.err</string>
  </dict>
</plist>" > $HOME/Library/LaunchAgents/com.renex.swapper.plist

