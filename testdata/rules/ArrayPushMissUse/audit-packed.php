<?php
$queue = ['first'];
$queue[<warning descr="The index is redundant here; use '[]' to append.">count</warning>($queue)] = 'next';
