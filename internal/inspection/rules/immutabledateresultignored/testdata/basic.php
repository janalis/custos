<?php
$day=new DateTimeImmutable('2025-03-10'); <warning descr="Assign the returned immutable date.">$day->modify('+1 day')</warning>; echo $day->format('c');
