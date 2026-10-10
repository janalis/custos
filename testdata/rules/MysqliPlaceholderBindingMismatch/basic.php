<?php
$db=new mysqli();$s=$db->prepare("SELECT ? + ?");<warning descr="Bind one value for each SQL placeholder.">$s->bind_param("i",$id)</warning>;
