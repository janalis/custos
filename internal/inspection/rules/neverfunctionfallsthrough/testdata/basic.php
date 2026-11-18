<?php
function halt(bool $stop): <error descr="Terminate every path in the never function.">never</error> { if ($stop) { exit; } }

// A loop that can end through its condition still reaches the closing brace.
function drain(array $queue): <error descr="Terminate every path in the never function.">never</error> { do { array_pop($queue); } while ($queue); }
function once(): <error descr="Terminate every path in the never function.">never</error> { do { usleep(1000); } while (false); }
function skip(): <error descr="Terminate every path in the never function.">never</error> { while (0) { usleep(1000); } }
function count3(int $i): <error descr="Terminate every path in the never function.">never</error> { for (; true, $i < 3; $i++) { usleep(1000); } }
function ready(): <error descr="Terminate every path in the never function.">never</error> { while (!is_file(__FILE__ . '.lock')) { usleep(1000); } }
