  func goroutine1() {
      m.Lock()
-     ch <- request //blocks  // HL
+     select {  // HL
+         case ch <- request: // HL
+         default: // HL
      }
      m.Unlock()
  }

  func goroutine2() {
      for {
          m.Lock() //blocks
          m.Unlock()
          request <- ch
      }
  }
