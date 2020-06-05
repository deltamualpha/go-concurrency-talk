

  for i := 17; i <= 21; i++ { // write
-     go func() { // HL
+     go func(i int) { // HL
          apiVersion := fmt.Sprintf("v1.%d", i)
-     }() // HL
+     }(i) // HL
  }