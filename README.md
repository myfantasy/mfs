# mfs
golang mfs - additional sync structs ant methods  

## mfs.RWTMutex - Read Write Try Mutex
mfs.RWTMutex is like sync.RWMutex  
``` golang
mx := mfs.RWTMutex{}
mx.Lock()
// DO Something
mx.Unlock()
```

and mfs.RWTMutex contains TryLock
``` golang
// ctx - sime context
ctx := context.Background()
mx := mfs.RWTMutex{}
isLocked := mx.TryLock(ctx)
if isLocked {
    // DO Something
    mx.Unlock()
} else {
    // DO Something else
}
```

## mfs.PMutex - Read Write Try Mutex with change priority (Promote and Reduce)
mfs.PMutex is like sync.RWMutex  
mfs.PMutex is like mfs.RWTMutex  

### Reduce - lock mutex from Lock to RLock
``` golang
mx := mfs.RWTMutex{}
mx.Lock()
// DO Something
mx.Reduce()
// Now mutex like RLock()
// DO Something
mx.RUnlock()
```

### Promote - lock mutex from RLock to Lock

``` golang
mx := mfs.RWTMutex{}
mx.RLock()
// DO Something
mx.Promote()
// Now mutex like Lock()
// DO Something
mx.Unlock()
```

### TryPromote - lock mutex from RLock to Lock
``` golang
// ctx is some context
mx := mfs.RWTMutex{}
mx.RLock()
// DO Something
isLocked := mx.TryPromote()
if isLocked {
    // Now mutex like Lock()
    // DO Something
    mx.Unlock()
} else {
    // Mutex is not locked (like call mx.RUnlock())
    // DO Something else
}
```

## F methods (PromoteF, TryPromoteF, LockF, TryLockF, RLockF, RTryLockF)
It should be locked befor regular methods like Lock or Promote are

## Me
Ilya Shcherbina  
+7(903)192-4239  
sch@myfantasy.ru  