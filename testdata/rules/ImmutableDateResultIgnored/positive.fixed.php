<?php
$day=new DateTimeImmutable('2025-03-10'); $day = $day->modify('+1 day'); echo $day->format('c');
