<?php
function bad(PDOStatement $s){<warning descr="Use associative fetch mode before serializing rows.">json_encode($s->fetchAll(PDO::FETCH_BOTH))</warning>;}
