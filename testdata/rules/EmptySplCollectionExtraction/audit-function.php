<?php
function f(){$q=new SplQueue();$q->count();$q->isEmpty();<error descr="Check that the collection contains an element.">$q->top()</error>;}
