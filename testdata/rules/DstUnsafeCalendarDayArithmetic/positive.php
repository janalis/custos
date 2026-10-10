<?php
$day=new DateTimeImmutable('2025-03-30 00:00:00',new DateTimeZone('Europe/Paris')); $next=<warning descr="Use calendar arithmetic for the next local day.">$day->getTimestamp()+86400</warning>;
