  func (p *peer) send(d []byte) error {
      p.mu.Lock()
      defer p.mu.Unlock()
      switch p.status {
      case idlePeer:
          if p.inflight.Get() > maxInflight {
              return fmt.Errorf("reach max idle")
          }
+         p.wg.Add(1) // HL
          go func() {
-             p.wg.Add(1) // HL
              p.post(d)
              p.wg.Done()
          }()
      return nil
  }
  func (p *peer) stop() {
      p.mu.Lock()
      if p.status == participantPeer {
          close(p.queue)
      }
      p.status = stoppedPeer
      p.mu.Unlock()
      p.wg.Wait()
  }