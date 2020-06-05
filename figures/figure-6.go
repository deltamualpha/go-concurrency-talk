- hctx, hcancel := context.WithCancel(ctx) // HL
+ var hctx context.Context // HL
+ var hcancel context.CancelFunc // HL
  if c.headerTimeout > 0 {
      hctx, hcancel = context.WithTimeout(ctx, c.headerTimeout)
+ } else { // HL
+     hctx, hcancel = context.WithCancel(ctx) // HL
  }
  defer hcancel()