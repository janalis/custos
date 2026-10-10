<?php
$start=new DateTimeImmutable(); $end=new DateTimeImmutable(); $days=$start->diff($end)->format('%a');
