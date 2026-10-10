package semanticquery

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestExpansionLanguageContracts(t *testing.T) {
	cases := []struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{
		{"InterfaceInstantiation", "<?php\ninterface Repository {} new Repository();\n", 1, []syntax.NodeKind{syntax.KNew}, CheckInterfaceInstantiation},
		{"InterfaceInstantiation", "<?php\ninterface Repository {} class DiskRepository implements Repository {} new DiskRepository();\n", 0, []syntax.NodeKind{syntax.KNew}, CheckInterfaceInstantiation},
		{"InterfaceInstantiation", "<?php interface X{} $name='X';new $name();", 1, []syntax.NodeKind{syntax.KNew}, CheckInterfaceInstantiation},
		{"InterfaceInstantiation", "<?php new Missing();", 0, []syntax.NodeKind{syntax.KNew}, CheckInterfaceInstantiation},
		{"InterfaceInstantiation", "<?php $name=$unknown;new $name();", 0, []syntax.NodeKind{syntax.KNew}, CheckInterfaceInstantiation},
		{"InterfaceInstantiation", "<?php trait T{}", 0, []syntax.NodeKind{syntax.KNew}, CheckInterfaceInstantiation},
		{"InterfaceInstantiation", "<?php if(false){\ninterface Repository {} new Repository();\n}", 0, []syntax.NodeKind{syntax.KNew}, CheckInterfaceInstantiation},
		{"InterfaceInstantiation", "<?php \ninterface Repository {} new Repository();\n $broken = ;", 0, []syntax.NodeKind{syntax.KNew}, CheckInterfaceInstantiation},
		{"CatchCannotHandleKnownThrowable", "<?php\ntry { throw new Error('failed'); } catch (Exception $e) {}\n", 1, []syntax.NodeKind{syntax.KCatch}, CheckCatchCannotHandleKnownThrowable},
		{"CatchCannotHandleKnownThrowable", "<?php\ntry { throw new Error('failed'); } catch (Throwable $e) {}\n", 0, []syntax.NodeKind{syntax.KCatch}, CheckCatchCannotHandleKnownThrowable},
		{"CatchCannotHandleKnownThrowable", "<?php try{throw new TypeError('x');}catch(Exception|RuntimeException $e){}", 1, []syntax.NodeKind{syntax.KCatch}, CheckCatchCannotHandleKnownThrowable},
		{"CatchCannotHandleKnownThrowable", "<?php try{throw new Exception('x');}catch(Error $e){}", 0, []syntax.NodeKind{syntax.KCatch}, CheckCatchCannotHandleKnownThrowable},
		{"CatchCannotHandleKnownThrowable", "<?php try{throw new Error('x');}catch(Missing $e){}", 0, []syntax.NodeKind{syntax.KCatch}, CheckCatchCannotHandleKnownThrowable},
		{"CatchCannotHandleKnownThrowable", "<?php try{unknown();throw new Error('x');}catch(Exception $e){}", 0, []syntax.NodeKind{syntax.KCatch}, CheckCatchCannotHandleKnownThrowable},
		{"CatchCannotHandleKnownThrowable", "<?php try{throw new Error('x');}catch(Exception $e){}finally{}", 0, []syntax.NodeKind{syntax.KCatch}, CheckCatchCannotHandleKnownThrowable},
		{"CatchCannotHandleKnownThrowable", "<?php try{$x=1;}catch(Exception $e){}", 0, []syntax.NodeKind{syntax.KCatch}, CheckCatchCannotHandleKnownThrowable},
		{"CatchCannotHandleKnownThrowable", "<?php try{throw $unknown;}catch(Exception $e){}", 0, []syntax.NodeKind{syntax.KCatch}, CheckCatchCannotHandleKnownThrowable},
		{"CatchCannotHandleKnownThrowable", "<?php class X extends Error{function __construct(){}}try{throw new X();}catch(Exception $e){}", 0, []syntax.NodeKind{syntax.KCatch}, CheckCatchCannotHandleKnownThrowable},
		{"CatchCannotHandleKnownThrowable", "<?php try{throw new Error(unknown());}catch(Exception $e){}", 0, []syntax.NodeKind{syntax.KCatch}, CheckCatchCannotHandleKnownThrowable},
		{"CatchCannotHandleKnownThrowable", "<?php try{throw new Error('x');}catch(Error $e){}catch(Exception $e){}", 0, []syntax.NodeKind{syntax.KCatch}, CheckCatchCannotHandleKnownThrowable},
		{"CatchCannotHandleKnownThrowable", "<?php if(false){\ntry { throw new Error('failed'); } catch (Exception $e) {}\n}", 0, []syntax.NodeKind{syntax.KCatch}, CheckCatchCannotHandleKnownThrowable},
		{"CatchCannotHandleKnownThrowable", "<?php \ntry { throw new Error('failed'); } catch (Exception $e) {}\n $broken = ;", 0, []syntax.NodeKind{syntax.KCatch}, CheckCatchCannotHandleKnownThrowable},
		{"NonVoidFunctionFallsThrough", "<?php\nfunction countItems(bool $ready): int { if ($ready) return 7; }\n", 1, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"NonVoidFunctionFallsThrough", "<?php\nfunction countItems(bool $ready): int { if ($ready) return 7; return 0; }\n", 0, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"NonVoidFunctionFallsThrough", "<?php function f():int{}", 1, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"NonVoidFunctionFallsThrough", "<?php function f():void{}", 0, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"NonVoidFunctionFallsThrough", "<?php function f():mixed{}", 0, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"NonVoidFunctionFallsThrough", "<?php function f():never{throw new Error();}", 0, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"NonVoidFunctionFallsThrough", "<?php function f(){}", 0, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"NonVoidFunctionFallsThrough", "<?php interface X{function f():int;}", 0, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"NonVoidFunctionFallsThrough", "<?php $f=function():int{};", 1, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"NonVoidFunctionFallsThrough", "<?php class X{function f():int{if(false){return 1;}}}", 1, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"NonVoidFunctionFallsThrough", "<?php function f():int{if(true){return 1;}}", 0, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"NonVoidFunctionFallsThrough", "<?php function f():int{if(false){return 1;}else{return 2;}}", 0, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"NonVoidFunctionFallsThrough", "<?php function f():int{if($b){return 1;}else{$x=2;}}", 1, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"NonVoidFunctionFallsThrough", "<?php function f():int{if($b){return 1;}elseif($c){return 2;}}", 0, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"NonVoidFunctionFallsThrough", "<?php function f():int{unknown();}", 0, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"NonVoidFunctionFallsThrough", "<?php function f():int{if(unknown()){return 1;}}", 0, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"NonVoidFunctionFallsThrough", "<?php function f():int{while($b){return 1;}}", 0, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"NonVoidFunctionFallsThrough", "<?php function f():int{throw new Error();}", 0, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"NonVoidFunctionFallsThrough", "<?php function f():int{yield 1;}", 0, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"NonVoidFunctionFallsThrough", "<?php function f():int{echo 'x';;}", 1, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"NonVoidFunctionFallsThrough", "<?php if(false){\nfunction countItems(bool $ready): int { if ($ready) return 7; }\n}", 0, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"NonVoidFunctionFallsThrough", "<?php \nfunction countItems(bool $ready): int { if ($ready) return 7; }\n $broken = ;", 0, []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}, CheckNonVoidFunctionFallsThrough},
		{"AssertionContainsRequiredSideEffect", "<?php\nassert(($result = 7) > 0); echo $result;\n", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect},
		{"AssertionContainsRequiredSideEffect", "<?php\n$result = 7; assert($result > 0); echo $result;\n", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect},
		{"AssertionContainsRequiredSideEffect", "<?php function f(){assert(($x=1)>0);echo $x;}", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect},
		{"AssertionContainsRequiredSideEffect", "<?php function f($x){assert(($x=1)>0);echo $x;}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect},
		{"AssertionContainsRequiredSideEffect", "<?php assert(($a[0]=1)>0);echo $a;", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect},
		{"AssertionContainsRequiredSideEffect", "<?php assert(($x=1)>0);$x=2;echo $x;", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect},
		{"AssertionContainsRequiredSideEffect", "<?php assert(($x=1)>0);echo $x;$x=2;", 1, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect},
		{"AssertionContainsRequiredSideEffect", "<?php echo $x;assert(($x=1)>0);echo $x;", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect},
		{"AssertionContainsRequiredSideEffect", "<?php function f(){global $x;assert(($x=1)>0);echo $x;}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect},
		{"AssertionContainsRequiredSideEffect", "<?php function f(){static $x;assert(($x=1)>0);echo $x;}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect},
		{"AssertionContainsRequiredSideEffect", "<?php assert(($x=1)>0);isset($x);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect},
		{"AssertionContainsRequiredSideEffect", "<?php assert(($x=1)>0);empty($x);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect},
		{"AssertionContainsRequiredSideEffect", "<?php assert(($x=1)>0);$y=$x??2;", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect},
		{"AssertionContainsRequiredSideEffect", "<?php assert(($x=1)>0);if($b){echo $x;}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect},
		{"AssertionContainsRequiredSideEffect", "<?php assert(($x=1)>0);$f=function(){echo $x;};", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect},
		{"AssertionContainsRequiredSideEffect", "<?php assert((function(){$x=1;return true;})());echo $x;", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect},
		{"AssertionContainsRequiredSideEffect", "<?php assert(($x=&$y)>0);echo $x;", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect},
		{"AssertionContainsRequiredSideEffect", "<?php assert(($x=1)>0);", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect},
		{"AssertionContainsRequiredSideEffect", "<?php if(false){\nassert(($result = 7) > 0); echo $result;\n}", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect},
		{"AssertionContainsRequiredSideEffect", "<?php \nassert(($result = 7) > 0); echo $result;\n $broken = ;", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect},
		{"FiberStartedTwice", "<?php\n$f = new Fiber(fn() => Fiber::suspend()); $f->start(); $f->start();\n", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberStartedTwice},
		{"FiberStartedTwice", "<?php\n$f = new Fiber(fn() => Fiber::suspend()); $f->start(); $f->resume();\n", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberStartedTwice},
		{"FiberStartedTwice", "<?php $f=new Fiber(fn()=>1);if($b){$f->start();}$f->start();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberStartedTwice},
		{"FiberStartedTwice", "<?php $f=new Fiber(fn()=>1);$f->start();$f=new Fiber(fn()=>1);$f->start();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberStartedTwice},
		{"FiberStartedTwice", "<?php $f=new Fiber(fn()=>1);$f->start();unknown($f);$f->start();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberStartedTwice},
		{"FiberStartedTwice", "<?php function f(Fiber $f){$f->start();$f->start();}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberStartedTwice},
		{"FiberStartedTwice", "<?php if(false){\n$f = new Fiber(fn() => Fiber::suspend()); $f->start(); $f->start();\n}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberStartedTwice},
		{"FiberStartedTwice", "<?php \n$f = new Fiber(fn() => Fiber::suspend()); $f->start(); $f->start();\n $broken = ;", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberStartedTwice},
		{"FiberResumedAfterTermination", "<?php\n$f = new Fiber(fn() => 7); $f->start(); $f->resume();\n", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberResumedAfterTermination},
		{"FiberResumedAfterTermination", "<?php\n$f = new Fiber(fn() => Fiber::suspend()); $f->start(); $f->resume();\n", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberResumedAfterTermination},
		{"FiberResumedAfterTermination", "<?php $f=new Fiber(function(){return 1;});$f->start();$f->resume();", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberResumedAfterTermination},
		{"FiberResumedAfterTermination", "<?php $f=new Fiber(function(){return null;});$f->start();$f->resume();", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberResumedAfterTermination},
		{"FiberResumedAfterTermination", "<?php $f=new Fiber(function(){});$f->start();$f->resume();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberResumedAfterTermination},
		{"FiberResumedAfterTermination", "<?php $f=new Fiber(fn()=>unknown());$f->start();$f->resume();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberResumedAfterTermination},
		{"FiberResumedAfterTermination", "<?php $f=new Fiber(fn()=>1);$f->resume();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberResumedAfterTermination},
		{"FiberResumedAfterTermination", "<?php $f=new Fiber($cb);$f->start();$f->resume();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberResumedAfterTermination},
		{"FiberResumedAfterTermination", "<?php if(false){\n$f = new Fiber(fn() => 7); $f->start(); $f->resume();\n}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberResumedAfterTermination},
		{"FiberResumedAfterTermination", "<?php \n$f = new Fiber(fn() => 7); $f->start(); $f->resume();\n $broken = ;", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberResumedAfterTermination},
		{"FiberReturnBeforeTermination", "<?php\n$f = new Fiber(fn() => Fiber::suspend()); $f->start(); $f->getReturn();\n", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberReturnBeforeTermination},
		{"FiberReturnBeforeTermination", "<?php\n$f = new Fiber(fn() => 7); $f->start(); $f->getReturn();\n", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberReturnBeforeTermination},
		{"FiberReturnBeforeTermination", "<?php $f=new Fiber(function(){Fiber::suspend();return 1;});$f->start();$f->getReturn();", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberReturnBeforeTermination},
		{"FiberReturnBeforeTermination", "<?php $f=new Fiber(fn()=>Fiber::suspend());$f->start();$f->resume();$f->getReturn();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberReturnBeforeTermination},
		{"FiberReturnBeforeTermination", "<?php $f=new Fiber(function(){});$f->start();$f->getReturn();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberReturnBeforeTermination},
		{"FiberReturnBeforeTermination", "<?php $f=new Fiber(function(){return 1;});$f->start();$f->getReturn();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberReturnBeforeTermination},
		{"FiberReturnBeforeTermination", "<?php $f=new Fiber(fn()=>Fiber::suspend());$f->getReturn();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberReturnBeforeTermination},
		{"FiberReturnBeforeTermination", "<?php if(false){\n$f = new Fiber(fn() => Fiber::suspend()); $f->start(); $f->getReturn();\n}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberReturnBeforeTermination},
		{"FiberReturnBeforeTermination", "<?php \n$f = new Fiber(fn() => Fiber::suspend()); $f->start(); $f->getReturn();\n $broken = ;", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberReturnBeforeTermination},
		{"FiberSuspendOutsideFiber", "<?php\nif (Fiber::getCurrent() === null) { Fiber::suspend(); }\n", 1, []syntax.NodeKind{syntax.KStaticCall}, CheckFiberSuspendOutsideFiber},
		{"FiberSuspendOutsideFiber", "<?php\nif (Fiber::getCurrent() !== null) { Fiber::suspend(); }\n", 0, []syntax.NodeKind{syntax.KStaticCall}, CheckFiberSuspendOutsideFiber},
		{"FiberSuspendOutsideFiber", "<?php if(null===Fiber::getCurrent()){Fiber::suspend();}", 1, []syntax.NodeKind{syntax.KStaticCall}, CheckFiberSuspendOutsideFiber},
		{"FiberSuspendOutsideFiber", "<?php Fiber::suspend();", 0, []syntax.NodeKind{syntax.KStaticCall}, CheckFiberSuspendOutsideFiber},
		{"FiberSuspendOutsideFiber", "<?php if($b){Fiber::suspend();}", 0, []syntax.NodeKind{syntax.KStaticCall}, CheckFiberSuspendOutsideFiber},
		{"FiberSuspendOutsideFiber", "<?php if(Fiber::getCurrent()===null){$f=fn()=>Fiber::suspend();}", 0, []syntax.NodeKind{syntax.KStaticCall}, CheckFiberSuspendOutsideFiber},
		{"FiberSuspendOutsideFiber", "<?php if(Fiber::getCurrent()===null){}else{Fiber::suspend();}", 0, []syntax.NodeKind{syntax.KStaticCall}, CheckFiberSuspendOutsideFiber},
		{"FiberSuspendOutsideFiber", "<?php if(false){\nif (Fiber::getCurrent() === null) { Fiber::suspend(); }\n}", 0, []syntax.NodeKind{syntax.KStaticCall}, CheckFiberSuspendOutsideFiber},
		{"FiberSuspendOutsideFiber", "<?php \nif (Fiber::getCurrent() === null) { Fiber::suspend(); }\n $broken = ;", 0, []syntax.NodeKind{syntax.KStaticCall}, CheckFiberSuspendOutsideFiber},
		{"GeneratorRewindAfterAdvance", "<?php\n$g = (function() { yield 7; yield 11; })(); $g->next(); $g->rewind();\n", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckGeneratorRewindAfterAdvance},
		{"GeneratorRewindAfterAdvance", "<?php\n$g = (function() { yield 7; yield 11; })(); $g->rewind();\n", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckGeneratorRewindAfterAdvance},
		{"GeneratorRewindAfterAdvance", "<?php function values(){yield 1;yield 2;}$g=values();$g->send(3);$g->rewind();", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckGeneratorRewindAfterAdvance},
		{"GeneratorRewindAfterAdvance", "<?php $g=(function(){yield 1;})();$g->next();$g->rewind();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckGeneratorRewindAfterAdvance},
		{"GeneratorRewindAfterAdvance", "<?php $g=(function(){if($b){yield 1;}yield 2;})();$g->next();$g->rewind();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckGeneratorRewindAfterAdvance},
		{"GeneratorRewindAfterAdvance", "<?php $g=(function(){yield 1;return;})();$g->next();$g->rewind();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckGeneratorRewindAfterAdvance},
		{"GeneratorRewindAfterAdvance", "<?php $g=unknown();$g->next();$g->rewind();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckGeneratorRewindAfterAdvance},
		{"GeneratorRewindAfterAdvance", "<?php $g=$unknown;$g->rewind();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckGeneratorRewindAfterAdvance},
		{"GeneratorRewindAfterAdvance", "<?php $g=(function(){yield 1;yield 2;})();$g->next();$g->{$name}();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckGeneratorRewindAfterAdvance},
		{"GeneratorRewindAfterAdvance", "<?php if(false){\n$g = (function() { yield 7; yield 11; })(); $g->next(); $g->rewind();\n}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckGeneratorRewindAfterAdvance},
		{"GeneratorRewindAfterAdvance", "<?php \n$g = (function() { yield 7; yield 11; })(); $g->next(); $g->rewind();\n $broken = ;", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckGeneratorRewindAfterAdvance},
		{"IteratorCountChangesRequiredPosition", "<?php\n$it = new ArrayIterator([7, 11]); iterator_count($it); echo $it->current();\n", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckIteratorCountChangesRequiredPosition},
		{"IteratorCountChangesRequiredPosition", "<?php\n$it = new ArrayIterator([7, 11]); iterator_count($it); $it->rewind(); echo $it->current();\n", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckIteratorCountChangesRequiredPosition},
		{"IteratorCountChangesRequiredPosition", "<?php $it=new ArrayIterator([1,2]);iterator_count($it);echo $it->key();", 1, []syntax.NodeKind{syntax.KMethodCall}, CheckIteratorCountChangesRequiredPosition},
		{"IteratorCountChangesRequiredPosition", "<?php $it=new ArrayIterator([1,2]);iterator_count($it);$it->seek(0);echo $it->current();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckIteratorCountChangesRequiredPosition},
		{"IteratorCountChangesRequiredPosition", "<?php $it=new ArrayIterator($items);iterator_count($it);echo $it->current();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckIteratorCountChangesRequiredPosition},
		{"IteratorCountChangesRequiredPosition", "<?php function f(ArrayIterator $it){iterator_count($it);echo $it->current();}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckIteratorCountChangesRequiredPosition},
		{"IteratorCountChangesRequiredPosition", "<?php $it=new ArrayIterator([1,2]);if($b){iterator_count($it);}echo $it->current();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckIteratorCountChangesRequiredPosition},
		{"IteratorCountChangesRequiredPosition", "<?php if(false){\n$it = new ArrayIterator([7, 11]); iterator_count($it); echo $it->current();\n}", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckIteratorCountChangesRequiredPosition},
		{"IteratorCountChangesRequiredPosition", "<?php \n$it = new ArrayIterator([7, 11]); iterator_count($it); echo $it->current();\n $broken = ;", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckIteratorCountChangesRequiredPosition},
	}
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"CatchCannotHandleKnownThrowable", "<?php try{return;}catch(Exception $e){}", 0, []syntax.NodeKind{syntax.KCatch}, CheckCatchCannotHandleKnownThrowable})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"CatchCannotHandleKnownThrowable", "<?php try{throw new Error;}catch(Exception $e){}", 1, []syntax.NodeKind{syntax.KCatch}, CheckCatchCannotHandleKnownThrowable})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"NonVoidFunctionFallsThrough", "<?php function f():int{$x=fn()=>unknown();}", 1, []syntax.NodeKind{syntax.KFunction}, CheckNonVoidFunctionFallsThrough})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"GeneratorRewindAfterAdvance", "<?php $g=(fn()=>1)();$g->rewind();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckGeneratorRewindAfterAdvance})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"GeneratorRewindAfterAdvance", "<?php $g=(function(){$x=1;yield 2;yield 3;})();$g->rewind();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckGeneratorRewindAfterAdvance})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"AssertionContainsRequiredSideEffect", "<?php assert(($x=1)>0);unset($x);echo $x;", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"AssertionContainsRequiredSideEffect", "<?php assert(($x=1)>0);$alias=&$x;echo $x;", 0, []syntax.NodeKind{syntax.KFuncCall}, CheckAssertionContainsRequiredSideEffect})

	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"NonVoidFunctionFallsThrough", "<?php function f():int{if(1){return 1;}}", 0, []syntax.NodeKind{syntax.KFunction}, CheckNonVoidFunctionFallsThrough})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"NonVoidFunctionFallsThrough", "<?php function f():int{if(1===1){return 1;}}", 0, []syntax.NodeKind{syntax.KFunction}, CheckNonVoidFunctionFallsThrough})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"NonVoidFunctionFallsThrough", "<?php function f($x):int{if($x===$x){return 1;}}", 0, []syntax.NodeKind{syntax.KFunction}, CheckNonVoidFunctionFallsThrough})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"NonVoidFunctionFallsThrough", "<?php function f($x):int{if($x>0){return 1;}}", 1, []syntax.NodeKind{syntax.KFunction}, CheckNonVoidFunctionFallsThrough})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"NonVoidFunctionFallsThrough", "<?php function f($x,$y):int{if($x&&$y){return 1;}}", 0, []syntax.NodeKind{syntax.KFunction}, CheckNonVoidFunctionFallsThrough})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"NonVoidFunctionFallsThrough", "<?php function f($x):int{if(!$x){return 1;}}", 0, []syntax.NodeKind{syntax.KFunction}, CheckNonVoidFunctionFallsThrough})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"FiberResumedAfterTermination", "<?php $f=new Fiber(fn($x)=>1);$f->start();$f->resume();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberResumedAfterTermination})
	cases = append(cases, struct {
		id, src string
		want    int
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{"FiberReturnBeforeTermination", "<?php $f=new Fiber(fn($x)=>Fiber::suspend());$f->start();$f->getReturn();", 0, []syntax.NodeKind{syntax.KMethodCall}, CheckFiberReturnBeforeTermination})

	for i, tc := range cases {
		t.Run(tc.id+"/"+fmt.Sprint(i), func(t *testing.T) {
			p := expansionCryptoDBProbe{id: tc.id, kinds: tc.kinds, check: func(ctx *analysis.Context, n syntax.Node) { tc.check(ctx, n, "finding") }}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("case.php", []byte(tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v\n%s", len(got), tc.want, got, tc.src)
			}
		})
	}
}

func TestExpansionLanguageBudget(t *testing.T) {
	for _, tc := range []struct {
		id, src string
		kind    syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{
		{"NonVoidFunctionFallsThrough", "<?php function f():int{" + strings.Repeat(";", 513) + "}", syntax.KFunction, CheckNonVoidFunctionFallsThrough},
		{"AssertionContainsRequiredSideEffect", "<?php assert(($x=1)>0);" + strings.Repeat("$y=1;", 513) + "echo $x;", syntax.KFuncCall, CheckAssertionContainsRequiredSideEffect},
	} {
		p := expansionCryptoDBProbe{id: tc.id, kinds: []syntax.NodeKind{tc.kind}, check: func(ctx *analysis.Context, n syntax.Node) { tc.check(ctx, n, "finding") }}
		e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}})
		if err != nil {
			t.Fatal(err)
		}
		if got := e.Analyze(syntax.Parse("case.php", []byte(tc.src), syntax.Options{})); len(got) != 0 {
			t.Fatal(got)
		}
	}
}

func TestExpansionLanguageMissingExpression(t *testing.T) {
	if expansionLanguageMayCall(nil) {
		t.Fatal("missing expression cannot contain a call")
	}
}
