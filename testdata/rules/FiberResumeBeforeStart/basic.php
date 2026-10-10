<?php
$f=new Fiber(fn()=>1); <warning descr="Start the fiber before resuming it.">$f->resume()</warning>;
