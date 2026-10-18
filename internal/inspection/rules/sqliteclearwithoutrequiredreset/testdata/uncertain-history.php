<?php
$db=new SQLite3(':memory:');$s=$db->prepare('SELECT :value');$r=$s->execute();if($flag){$s->clear();}if($r->fetchArray()){$s->bindValue(':value',1);$s->execute();}
