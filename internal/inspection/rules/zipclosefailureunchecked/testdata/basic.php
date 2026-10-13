<?php
function archive($p){$z=new ZipArchive();if($z->open($p,ZipArchive::CREATE)!==true){return false;}$z->addFromString("a.txt","a");<warning descr="Check archive finalization before returning success.">$z->close()</warning>;return true;}
