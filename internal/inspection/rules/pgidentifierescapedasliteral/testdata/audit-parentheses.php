<?php
$name=(<warning descr="Escape PostgreSQL identifiers with the identifier API.">pg_escape_literal($db,"items")</warning>);pg_query($db,"SELECT * FROM ".$name);echo $name;
