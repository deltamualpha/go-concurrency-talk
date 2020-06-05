  var group sync.WaitGroup
  group.Add(len(pm.plugins))
  for _, p := range pm.plugins {
	  go func(p *plugin) {
		  defer group.Done()
		  if err := pm.restorePlugin(p); err != nil {
			  logrus.Errorf("Error restoring plugin '%s': %s", p.Name(), err)
			  return
		  }
		  pm.Lock()
		  pm.nameToID[p.Name()] = p.PluginObj.ID
		  requiresManualRestore := !pm.liveRestore && p.PluginObj.Active
		  pm.Unlock()
		  if requiresManualRestore {
			  // if liveRestore is not enabled, the plugin will be stopped now so we should enable it
			  if err := pm.enable(p); err != nil {
				  logrus.Errorf("Error enabling plugin '%s': %s", p.Name(), err)
			  }
		  }
	  }(p)
-     group.Wait() // HL
  }
+ group.Wait() // HL
  return pm.save()
