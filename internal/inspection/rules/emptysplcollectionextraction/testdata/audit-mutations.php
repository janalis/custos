<?php
$q=new SplQueue();$q[]=1;echo $q->dequeue();
$r=new SplQueue();$r->offsetSet(null,1);echo $r->dequeue();
$t=new SplQueue();if($flag){$t->enqueue(1);}echo $t->dequeue();
$u=new SplQueue();$u->$method(1);echo $u->dequeue();
