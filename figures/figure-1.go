// BASIC OMIT

func finishReq(timeout time.Duration) r ob {
	- ch := make(chan ob) // HL
	+ ch := make(chan ob, 1) // HL
	go func() {
		result := fn()
		ch <- result // block
	}()
	select {
	case result = <- ch:
		return result
	case <- time.After(timeout):
		return nil
	}
}

// ENDBASIC OMIT

// FULL OMIT

func finishRequest(timeout time.Duration, fn resultFunc) (result runtime.Object, err error) {
	- ch := make(chan runtime.Object) // HL
	- errCh := make(chan error) // HL
	+ ch := make(chan runtime.Object, 1) // HL
	+ errCh := make(chan error, 1) // HL
	go func() {
		if result, err := fn(); err != nil {
			errCh <- err
		} else {
			ch <- result
		}
	}()
	select {
	case result = <-ch:
		if status, ok := result.(*api.Status); ok {
			return nil, errors.FromObject(status)
		}
		return result, nil
	case err = <-errCh:
		return nil, err
	case <-time.After(timeout):
		return nil, errors.NewTimeoutError("request did not complete within allowed duration")
	}
}

// ENDFULL OMIT