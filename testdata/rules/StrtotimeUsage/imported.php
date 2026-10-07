<?php
namespace App;

use function Vendor\Clock\time;

$a = <warning descr="Call time() instead of parsing 'now'.">strtotime('now')</warning>;
