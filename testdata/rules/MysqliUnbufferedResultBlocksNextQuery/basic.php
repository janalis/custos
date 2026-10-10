<?php
$db=new mysqli();$r=$db->query("SELECT id FROM users",MYSQLI_USE_RESULT);<warning descr="Consume or free the unbuffered result before another query.">$db->query("SELECT 1")</warning>;
