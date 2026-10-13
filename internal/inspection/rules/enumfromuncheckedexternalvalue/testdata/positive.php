<?php
enum Phase:string {case Ready='ready';} $p=<warning descr="Handle unknown enum backing values explicitly.">Phase::from($_GET['phase'])</warning>;
