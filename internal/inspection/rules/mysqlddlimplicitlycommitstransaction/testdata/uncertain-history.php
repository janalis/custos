<?php
$db=new PDO('mysql:host=localhost');$db->beginTransaction();if($flag){$db->exec('CREATE TABLE example (id INT)');}$db->rollBack();
