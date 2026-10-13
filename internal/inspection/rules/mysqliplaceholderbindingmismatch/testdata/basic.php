<?php
$db=new mysqli();$s=$db->prepare("SELECT ? + ?");<warning descr="Bind one value for each SQL placeholder.">$s->bind_param("i",$id)</warning>;
$c=$db->prepare("SELECT ? -- first ?\n + ? -- last");<warning descr="Bind one value for each SQL placeholder.">$c->bind_param("i",$id)</warning>;
$d=$db->prepare("SELECT ? --\n-- ?\n");<warning descr="Bind one value for each SQL placeholder.">$d->bind_param("ii",$id,$id)</warning>;
