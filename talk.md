# Understanding Real-World Concurrency Bugs in Go

## Section 1

> Go advocates for the usage of message passing as the means of inter-thread communication and provides several new concurrency mechanisms and libraries to ease multi-threading programming. It is important to understand the implication of these new proposals and the comparison of message passing and shared memory synchronization in terms of program errors, or bugs.

For "message passing", the paper cites "C. A. R. Hoare - Communicating Sequential Processes - Communications of the ACM, 21(8):666-677, 1978" - <https://www.cs.cmu.edu/~crary/819-f09/Hoare78.pdf>

With that citation, I started following a rabbit hole of lookups and citations that I ended up finding very interesting, so let's go exploring for a few minutes, shall we?

CSP is a general-purpose model for inter-process communication, and here "process" is being defined _very_ loosely. It centers on the concept of "inputs" and "outputs" between processes that can be linked together. The Hoare paper is the landmark in the field, and basically everyone cites it constantly. Erlang is based on this model. Linux pipes are based on this model; Russ Cox mentions this in his history of Bell Labs and CSP: <https://swtch.com/~rsc/thread/>

> Of course, the Unix pipe mechanism doesn't require the linear layout; only the shell syntax does. McIlroy reports toying with syntax for a shell with general plumbing early on but not liking the syntax enough to implement it. Later shells did support some restricted forms of non-linear pipelines. Rochkind's `2dsh` supports dags; Tom Duff's `rc` supports trees.

Multi-dimensional pipes! I can't even imagine the syntax. It'd probably look something like named pipes (which are already something of a bear to work with)?

Curiously, some of these diagrams remind me of the TIS-100, a game by Zachtronics, which, now that I think about it, more or less implements a variety of CSP using unbuffered communication between neighboring processes. <http://www.zachtronics.com/tis-100/>

How does this relate to Go?

The Go FAQ talks about the influence of CSP on Go: <https://golang.org/doc/faq#csp>

> One of the most successful models for providing high-level linguistic support for concurrency comes from Hoare's Communicating Sequential Processes, or CSP. Occam and Erlang are two well known languages that stem from CSP. Go's concurrency primitives derive from a different part of the family tree whose main contribution is the powerful notion of channels as first class objects. Experience with several earlier languages has shown that the CSP model fits well into a procedural language framework.

And Russ Cox, as mentioned previously, gives us another piece of the puzzle: Bell Labs.

Rob Pike: Squeak -> Newsqueak -> Plan 9 and Inferno (Alef, Limbo) -> go

It's almost certainly overly reductive to view this chain of languages through the lens of Rob Pike's involvement. Wikipedia explicitly names a number of collaborators on these projects at Bell Labs over the years... but Rob Pike is a constant.

Curiously, the first forays into this that I have documentation of are centered on _graphical user interfaces_ -- Squeak and Newsqueak are named as such because they interact with the mouse. I say "curiously" because Go has no GUI components in its standard library, and is definitely not put forward as a language that's _designed_ for GUI use. <https://swtch.com/~rsc/thread/cws.pdf>

> The channels that carry messages between processes in Newsqueak, and hence within the window system, are synchronous, bufferless channels as in Communicating Sequential Processes (CSP) or Squeak, but carry objects of a specific type, not just integers. To receive a message on a channel, say M, a process executes
>
> ```squeak
> mrcv = <-M
> ```
>
> and blocks until a different process executes
>
> ```squeak
> M<- = msend
> ```
>
> for the same channel. When both a sender and a receiver are ready, the value is transferred between the processes (here assigning the value of `msend` to `mrcv`), which then resume execution. Except for the syntax, this sequence is exactly as in CSP.

This syntax should look _very_ familar to anyone who's used go.

And to close the loop: Thomas Kappler went through and implemented all of the examples of CSP from the Hoare paper in go: <https://github.com/thomas11/csp/> (which is quite handy for the modern reader), and the Go blog has an early post called "Share Memory By Communicating" that shows how programmers can use message passing instead of mutexes and locks to build simpler, cleaner, less buggy parallel programs: <https://blog.golang.org/share-memory-by-communicating>

Sir Charles Antony Richard Hoare -- Tony to his friends! -- is still alive, by the way, in his 80s and living in the UK. <http://www.cs.ox.ac.uk/people/tony.hoare/>

CSP probably isn't even the thing he's most famous for: he also developed Quicksort and worked on Algol 60 (the granddaddy of the entire C branch of programming languages), and as part of that later apologized for introducing the concept of the "null reference" to computer science. <https://www.infoq.com/presentations/Null-References-The-Billion-Dollar-Mistake-Tony-Hoare/>

There's also a quote from him that I'm stealing off his wikipedia entry that I wanted to share with you all.

> Ten years ago, researchers into formal methods (and I was the most mistaken among them) predicted that the programming world would embrace with gratitude every assistance promised by formalisation to solve the problems of reliability that arise when programs get large and more safety-critical. Programs have now got very large and very critical – well beyond the scale which can be comfortably tackled by formal methods. There have been many problems and failures, but these have nearly always been attributable to inadequate analysis of requirements or inadequate management control. It has turned out that the world just does not suffer significantly from the kind of problem that our research was originally intended to solve.

### Back to the paper

So! That's a whirlwind tour of what message-passing and CSP are. With that background, the paper posits the following:

> A major design goal of Go is to improve traditional multi-threaded programming languages and make concurrent programming easier and less error-prone.

Whether this is _true_ or not is open to debate, and the authors here attempt to see if go _actually_ makes concurrent programming easier and less error-prone through emperical research on real codebases.

How do they do it?

>In this paper, we conduct the first empirical study on Go concurrency bugs using six open-source, production- grade Go applications: Docker and Kubernetes, two datacenter container systems, etcd, a distributed key-value store system, gRPC, an RPC library, and CockroachDB and BoltDB, two database systems.
>
> In total, we have studied 171 concurrency bugs in these applications. We analyzed the root causes of them, performed experiments to reproduce them, and examined their fixing patches. Finally, we tested them with two existing Go concurrency bug detectors (the only publicly available ones).

(Note that development work on BoltDB has since been frozen and the project forked to bbolt under the auspices of the ectd team: <https://github.com/etcd-io/bbolt>)

Their method here is not "we're going to go examine these code bases and find new bugs" but "we're going to go through these large go language projects' github repositories, examine all of their issues and merge requests, identify those that involve 'concurrency', and then examine causes and fixes".

```diff
func finishRequest(timeout time.Duration, fn resultFunc) (result runtime.Object, err error) {
- ch := make(chan runtime.Object)
- errCh := make(chan error)
+ // these channels need to be buffered to prevent the goroutine below from hanging indefinitely
+ // when the select statement reads something other than the one the goroutine sends on.
+ ch := make(chan runtime.Object, 1)
+ errCh := make(chan error, 1)
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
```

This is their first example of what they're talking about, a pretty simple bug that leaks goroutines in kubernetes. It's <https://github.com/kubernetes/kubernetes/pull/5316> in case you're curious; they've trimmed the code down to make it fit into their paper column, which _dramatically_ hurts readability in my opinion (their example isn't actually valid go code and obscures some of the actual mechanisms of the function), so I'm going to show the actual patch. A couple of notes here in case you're not familiar with go:

* An anonymous goroutine, here defined by `go func() {}`, inherits the scope of its parent -- `ch` inside that goroutine is the same as the one in the parent. That's true not only of named channels like that but any other variables, and important to keep in mind for some of the other examples, which depend on misuse of this fact.
* `select{}` will block until one of its cases becomes executable -- but if more than one case is "available", the runtime will _randomly_ pick a case to execute. Presumably, this randomization is to prevent programmers from becoming dependent on an implementation detail of how the runtime decides which branch to pick, but it is also a source of bugs. (Go also explicitly refuses to promise a stable iteration order over maps to prevent programmers from relying on an implementation detail that may not be the same across releases or hardware.)
* The _bug_ here is that, if the timeout branch is chosen by the runtime, the parent function returns and the child goroutine will _never_ be able to return, blocking forever on `ch <- result` or `errCh <- err`, because the original code had unbuffered channels (that is, channels that block writers until there's a reader available on the other end). Switching to buffered channels allows the child goroutine to push its message onto a channel and exit, even if the parent function has already returned `nil, error` and exited. Go's garbage collector will eventually reap any orphaned channels and any values that were written into them (as long as there are no other references to those values).

It's worth noting here the only way I was able to reconstruct this real code example from the figure in the paper is because the examples they used to create this paper are all online, in an Excel sheet (groan), with references to commit hashes and issue numbers: <https://github.com/system-pclub/go-concurrency-bugs>

## Section 2

We've talked about what message-passing and CSP are, but let's review locks and more "traditional" forms of concurrency management in the go context:

* Mutual Exclusion Lock a.k.a. `Mutex`: basic locks. Only one process can have the lock, any other calls to try and get the lock block until the lock is available again.

  > ```go
  > // SafeCounter is safe to use concurrently.
  > type SafeCounter struct {
  >   v   map[string]int
  >   mux sync.Mutex
  > }
  >
  > // Inc increments the counter for the given key.
  > func (c *SafeCounter) Inc(key string) {
  >   c.mux.Lock()
  >   // Lock so only one goroutine at a time can access the map.
  >   defer c.mux.Unlock()
  >   c.v[key]++
  > }
  > ```

* Read/Write Mutual Exclusion Lock a.k.a. `RWMutex`: many simultaneous readers, but only one writer, who blocks all other accesses until the lock is released. One wrinkle: a call to `Lock()` that's blocked takes priority over all other `RLock()` calls, to make sure that writer can break through a read-heavy mutex. This is a change from how this is implemented in other languages.

  > ```go
  > // SafeCounter is safe to use concurrently.
  > type SafeCounter struct {
  >   v   map[string]int
  >   mux sync.RWMutex
  > }
  >
  > // Inc increments the counter for the given key.
  > func (c *SafeCounter) Inc(key string) {
  >   c.mux.Lock()
  >   // Lock so only one goroutine at a time can access the map.
  >   defer c.mux.Unlock()
  >   c.v[key]++
  > }
  >
  > // Value returns the current value of the counter for the given key.
  > func (c *SafeCounter) Value(key string) int {
  >   c.mux.RLock()
  >   // Many readers can get this lock at once, but Lock() takes priority.
  >   defer c.mux.RUnlock()
  >   return c.v[key]
  > }
  > ```

* Conditional Variables a.k.a. `Cond`: built around a mutex, this adds `Wait()` and `Broadcast()`, which allow a process with the lock to temporarily give it up and wait for the next process to release it again, at which point the other process needs to broadcast and tell all waiting processes that they can resume again.

  > ```go
  > func main() {
  >     var v int
  >     m := sync.Mutex{}
  >     m.Lock() // 1. main process is owner of lock.
  >     c := sync.NewCond(&m)
  >     go func() {
  >         m.Lock() // 4. blocks until Wait(); goroutine then owns lock.
  >         v = 1
  >         c.Broadcast() // 5. let waiters know we're done.
  >         m.Unlock() // 6. goroutine has released lock.
  >     }()
  >     v = 0 // 2. do some initialization.
  >     c.Wait() // 3. Temporarily release the lock and block until Broadcast() is called.
  >     fmt.Println(v) // 7. prints "1".
  >     m.Unlock()
  > }
  > ```

  In theory, any process that's waiting should actually check to see if the state of the system that it's waiting on has actually been achieved before proceeding, since more than one process might be waiting and someone else might have gotten to the piece of shared state that we were waiting for first. Cond can be confusing because it's a semaphore for marking that some condition has been reached but does not in itself contain that condition, which can lead to race conditions if not used with care.

* `Once` is a go construct that provides that a function will only be called, well, once. Note that it doesn't matter _what_ function you give it; you can call the same `Do()` with different functions and it'll still only run the first one provided -- this is not the same as memoization, not least of which that `Do()` does not return anything.

  > ```go
  > func main() {
  >     var once sync.Once
  >     onceBody := func() {
  >         fmt.Println("Only once")
  >     }
  >     done := make(chan bool)
  >     for i := 0; i < 10; i++ {
  >         go func() {
  >             // prints "Only once" once, though Do() is called ten times.
  >             once.Do(onceBody)
  >             done <- true
  >         }()
  >     }
  >     for i := 0; i < 10; i++ {
  >         <-done // wait for all goroutines to finish.
  >     }
  >     fmt.Println("Complete")
  > }
  > ```

* `WaitGroup` is a construct to make that idiom of waiting for all goroutines to finish we saw in the last example easier: given a bunch of processes, block until they're all marked as complete before proceeding.

  > ```go
  > func main() {
  >     var wg sync.WaitGroup
  >     for i := 0; i < 10; i++ {
  >         wg.Add(1)
  >         go func(routine int) {
  >             defer wg.Done()
  >             fmt.Println(routine)
  >         }(i)
  >     }
  >     // wait for all goroutines to finish.
  >     fmt.Println("Started all routines")
  >     wg.Wait()
  >     fmt.Println("Done")
  > }
  > ```

The example for `Once` I used also points toward something interesting: I took that directly from the documentation for the go sync library (with some tweaks for clarity). Go doesn't treat concurrency models as an either-or proposition; the documentation freely mixes "traditional" concurrency primitives with channel-based ones.

## Section 3

Table 2 in the paper is interesting.

> Overall, the six applications use a large amount of goroutines.

Now, I don't know about you, but `2` in the case of BoltDB does not strike me as a large amount of goroutines.

> **Observation 1**: Goroutines are shorter but created more frequently than C (both statically and at runtime).

This follows from the stated goal of go to make goroutines "cheap", but I don't know if there's actually interesting results here, because...

> We found all threads in gRPC-C execute from the beginning to the end of the whole program (i.e., 100%)

This implies to me that the C threads and the goroutines are serving _different practical functions_. And while they may conceptually be the "same primitives" in some sense, does this result actually tell us anything? If the threads were being spun up in response to requests, as the goroutines presumably are, their relative execution times might be useful to compare, but as it is I don't know that they've actually collected enoughe evidence to make this observation.

Figure 2 and Figure 3 are mostly uncommented-on in the text of the paper... and it's fun to note that they're identical, just inverted: "all primitives" is just shared-memory plus message-passing. `etcd`'s status as an outlier here is interesting -- I wish there was some examination of these six projects in relation to each other, why they might be choosing shared-memory vs message-passing, but that's not what the paper's authors chose to focus on in their introduction. I think that'd be more revealing than comparing gRPC's go and C implementations!

> **Implication 1**: With heavier usages of goroutines and new types of concurrency primitives, Go programs may potentially introduce more concurrency bugs.

With this, they lay out the actual question they're interested in trying to answer in the paper.

## Section 4

> To collect concurrency bugs, we first filtered GitHub commit histories of the six applications by searching their commit logs for concurrency-related keywords, including “race”, “deadlock”, “synchronization”, “concurrency”, “lock”, “mutex”, “atomic”, “compete”, “context”, “once”, and “goroutine leak”. Some of these keywords are used in previous works to collect concurrency bugs in other languages. Some of them are related to new concurrency primitives or libraries introduced by Go, such as “once” and “context”. One of them, “goroutine leak”, is related to a special problem in Go. In total, we found 3211 distinct commits that match our search criteria.

It's worth noting that they're using slightly different categories than other papers on this topic: "blocking vs. non-blocking" instead of "deadlock vs non-deadlock". The main difference is "blocking" includes not just deadlocks, where all processes cannot move forward, but also places where just some _portion_ of the overall program is stuck waiting for a resource that can never be satisfied. Thinking back to the example given with the kubernetes bug -- changing from an unbuffered to a buffered channel -- we can say this definition captures "leaky" goroutines alongside full deadlocks.

> We then randomly sampled the filtered commits, identified commits that fix concurrency bugs, and manually studied them. Many bug-related commit logs also mention the corresponding bug reports, and we also study these reports for our bug analysis. We studied 171 concurrency bugs in total.

Why random sampling? I assume this is so they can get an "unbiased" look at the commits? But that also means that this is of course a subset of the actual bugs that might have been present over the lifetime of these applications and we're relying on the power of _statistics_ -- but are all classes of bugs evenly distributed? No way of knowing. The bug lifetime graph in Figure 4 implies that both classes of bugs share the same "time to discovery and fix" characteristics, which does lend some support to their findings.

> We only studied concurrency bugs that have been fixed. There could be other concurrency bugs that are rarely reproduced and are never fixed by developers. For some fixed concurrency bugs, there is too little information provided, making them hard to understand. We do not include these bugs in our study.

A noble admission of the limits of their own comprehension, but some data on what "class" they suspect these bugs would have fallen into would have been nice.

## Section 5

> Overall, we found that there are around 42% blocking bugs caused by errors in protecting shared memory, and 58% are caused by errors in message passing. Considering that shared memory primitives are used more frequently than message passing ones, message passing operations are even more likely to cause blocking bugs.

Which leads to...

> Observation 3: Contrary to the common belief that message passing is less error-prone, more blocking bugs in our studied Go applications are caused by wrong message passing than by wrong shared memory protection.

This is the big headline takeaway that the paper got some traction for, and it's definitely how I first heard about it. "Go's designers talk a big game about how much better and safer their language is because of channels, but _really_ it's worse!" Everyone loves to cover a story about how conventional wisdom is wrong! And the paper hit just as go really seemed to be taking off, mindshare-wise, among the broader programming community.

But if you look at the tables and charts in this section, I think this observation seems to be shaving the numbers a little thinly. Table 5 shows that, _overall_, shared memory misprotection contributed to more bugs than message passing, just a hair over 60% of analysed bugs. And referencing Table 4, shared memory is used _more often_ in general for concurrency; depending on the application, between 60%-80% of concurrency primitives are shared memory over message passing. And, of course, there's the fudge of "channel w/ other", which seems to stand for "channel operations and memory-sharing operations interleaved" -- why does the channel part get blamed there instead of the misuse of a mutex?

Finally, there's the distinction of "blocking vs nonblocking", which sounds like a difference in severity when you read it, but I would argue that there's no categorical way to assert that blocking bugs are "more important" or "more serious" than non-blocking bugs. "Yes, our goroutines never block, but the data race that's in there results in every checking account processed by the application to silently lose a penny on every transaction." Program incorrectness comes in many forms!

But enough about their conclusions, let's look at some of the examples. Again, I'll be showing the full patches when possible here instead of the trimmed-down cases from the paper.

This is the full context for Figure 5 in the paper.

```diff
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
-   group.Wait()
  }
+ group.Wait()
  return pm.save()
```

<https://github.com/moby/moby/pull/25384>. Super-simple bug, super-simple fix: create a WaitGroup, increment it up to wait for the number of plugins, then range over the plugins and fire off goroutines for each one, doing some registration and decrementing the group at the end of each goroutine. They even properly passed each round's plugin into the goroutine as a function argument instead of relying on the shared variable, which is a _very_ common bug... but then just put the `Wait()` in the wrong place, waiting at the end of each loop of the range instead of after the range was finished. Works just fine for the `0` and `1` plugin cases, blocks forever in the `1+n` case, waiting for calls to `Done()` that the `Wait()` itself is keeping from proceeding.

> Although conditional variable and thread group wait are both traditional concurrency techniques, we suspect Go’s new programming model to be one of the reasons why programmers made these concurrency bugs. For example, unlike `pthread_join` which is a function call that explicitly waits on the completion of (named) threads, `WaitGroup` is a variable that can be shared across goroutines and its `Wait` function implicitly waits for the `Done` function.

Interestingly, it wouldn't _block_ if they called `group.Add()` inside the range loop before starting a goroutine, but it also would probably still be a bug: you'd only get sequential processing of the goroutines, because the group would `Wait()` for each run to finish. Compare:

```go
package main

import (
  "fmt"
  "sync"
)

func main() {
  fmt.Println("begin")
  r := []int{1, 2, 3, 4, 5}
  var group sync.WaitGroup

  for _, p := range r {
    group.Add(1)
    go func(p int) {
      defer group.Done()
      fmt.Println(p)
    }(p)
    group.Wait()
    fmt.Println("done waiting")
  }
  fmt.Println("done")

  fmt.Println("begin")
  for _, p := range r {
    group.Add(1)
    go func(p int) {
      defer group.Done()
      fmt.Println(p)
    }(p)
  }
  group.Wait()
  fmt.Println("done waiting")
  fmt.Println("done")
}
```

It's hard to call this a big bug; I don't know that it was caused by some _misunderstanding_ of the concurrency primitive here. If you look at the PR where it was introduced, <https://github.com/moby/moby/pull/23580>, I read that patch and this resulting bug as an typo that snuck in because the programmer who wrote this restore feature only tested with 0 and 1 plugins. And, to be fair, that's probably partially the point. This is an obvious, easy mistake that's well-understood but also hard to catch when your eyes are skimming over the code for the _business logic_ pieces and not the _concurrency_ pieces.

> We now discuss blocking bugs caused by errors in message passing, which in the contrary of common belief are the main type of blocking bugs in our studied applications.

Oho! The stars of the show!

The PR here is <https://github.com/etcd-io/etcd/pull/3927>.

```diff
- hctx, hcancel := context.WithCancel(ctx)
+ var hctx context.Context
+ var hcancel context.CancelFunc
  if c.headerTimeout > 0 {
    hctx, hcancel = context.WithTimeout(ctx, c.headerTimeout)
+ } else {
+   hctx, hcancel = context.WithCancel(ctx)
  }
  defer hcancel()
```

This one is so simple they don't even have to reformat it to fit into the column (this is Figure 6).

> When combining with the usage of Go special libraries, the channel creation and goroutine blocking may be buried inside library calls. As shown in Figure 6, a new context object, hcancel, is created at line 1. A new goroutine is created at the same time, and messages can be sent to the new goroutine through the channel field of hcancel. If timeout is larger than 0 at line 4, another context object is created at line 5, and hcancel is pointing to the new object. After that, there is no way to send messages to or close the goroutine attached to the old object. The patch is to avoid creating the extra context object when timeout is larger than 0.

The original code created a cancellation context based on the passed context, then, if there was an explicit timeout set, set that same cancellation context variable to the timeout context instead. The actual channels being used here are hidden within the Context library, but suffice to say creating a context and then immediately overwriting your reference to it leaves a channel hanging around in the background. (The `hcancel` function can also be called in the case of a cancel, so the patched version is functionally identical to the buggy version.)

From the go documentation:

> The WithCancel, WithDeadline, and WithTimeout functions take a Context (the parent) and return a derived Context (the child) and a CancelFunc. Calling the CancelFunc cancels the child and its children, removes the parent's reference to the child, and stops any associated timers. Failing to call the CancelFunc leaks the child and its children until the parent is canceled or the timer fires. The go vet tool checks that CancelFuncs are used on all control-flow paths.

It's subtle! It's easy to get wrong because the concurrency primitives are hidden away from you. Go is garbage-collected, so you don't need to worry about an extraneous struct lifetime, but there is no such reaper for goroutines.

Let's look at Figure 7, which I was unable to find the original bug for:

```diff
  func goroutine1() {
    m.Lock()
-   ch <- request //blocks
+   select {
+     case ch <- request:
+     default:
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
```

`goroutine1` attempts to send the `request` onto the channel while holding the lock; `goroutine2` attempts to grab the lock, then read from the channel. Neither can proceed. The addition of the empty select lets `goroutine1` proceed past the send in the case of a block, unlocking the lock so that `goroutine2` can itself receive.

This is one of the more annoying examples in the paper because it's been chopped down to the point of incoherence in order to fit onto the page. Is there supposed to be a defer on the `m.Unlock()` in goroutine2? Why not just make the channel buffered?

It's also worth noting here that this does _change_ the behavior of the program: the select will now opt to skip sending the message on the blocked channel in the fixed version, and `goroutine2` will now block trying to read from the channel! Whether this is acceptable or not in the context of the original program is left as an exercise for the reader to discover; regardless, the explanation provided in the paper:

> The fix is to add a `select` with `default` branch for goroutine1 to make `ch` not blocking any more.

is insufficient to explain the patch shown.

Finally, we reach another Observation:

> Observation 5: All blocking bugs caused by message passing are related to Go’s new message passing semantics like channel. They can be difficult to detect especially when message passing operations are used together with other synchronization mechanisms.

This seems obvious when combined with the earlier chart that the vast majority of "message passage semantics" are implemented using channel. The bugs are in the code that's written, not code that isn't.

We'll save the detection discussion for the end; for now, onto the non-blocking bugs!

## Section 6

Figure 8 has probably bitten most go developers at one time or another. (The GitHub PR number provided in the text as an example of a "traditional bug" here is unrelated to the Docker bug used as an example in Figure 8, which is moby/moby#27037.)

```diff
  for i := 17; i <= 21; i++ { // write
-   go func() {
+   go func(i int) {
      apiVersion := fmt.Sprintf("v1.%d", i)
-   }()
+   }(i)
  }
```

Variables in anonymous goroutines are shared with the parent unless shadowed by a function argument. Here, the buggy version of this (test!) code uses the value of `i` at goroutine execution time, not creation time, and since goroutines can be arbitrarily scheduled, more often than not all executions of the test were run with `i` set to `21`. Properly explicitly passing the value of `i` with every loop fixes the problem (although I personally would rename the variable inside the goroutine, to prevent confusion for someone else who might encounter this code down the line). Shared variable access is shared memory!

Figure 9 is another subtle error:

```diff
  func (p *peer) send(d []byte) error {
    p.mu.Lock()
    defer p.mu.Unlock()

    switch p.status {
    case participantPeer:
      select {
      case p.queue <- d:
      default:
        return fmt.Errorf("reach max serving")
      }
    case idlePeer:
      if p.inflight.Get() > maxInflight {
        return fmt.Errorf("reach max idle")
      }
+     p.wg.Add(1)
      go func() {
-       p.wg.Add(1)
        p.post(d)
        p.wg.Done()
      }()
    case stoppedPeer:
      return fmt.Errorf("sender stopped")
    }
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
```

Why is this necessary? Again, it's because a goroutine can be arbitrarily scheduled. While the entire `p.send()` function is inside a critical section and has the lock, the anonymous goroutine created from inside it escapes the critical section and may run even after the lock has been released. As such, `p.send()` might run, spin off its goroutine, return, and then `p.stop()` might run, correctly grab the lock, check its status, and then call `Wait()`, all before `p.wg.Add(1)` gets a chance to execute. The `Wait()` won't block since the counter was never incremented, presumably then discarding the `p.post(p)` call inside the goroutine.

## Community reactions

<https://news.ycombinator.com/item?id=19280927>

<https://www.reddit.com/r/golang/comments/au938q/understanding_realworld_concurrency_bugs_in_go_pdf/>
