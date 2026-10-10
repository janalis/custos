<?php
$start=new DateTimeImmutable(); $end=new DateTimeImmutable(); $days=<warning descr="Use total days instead of the day component.">$start->diff($end)->format('%d')</warning>;
