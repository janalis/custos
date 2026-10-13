<?php
$day=new DateTimeImmutable(); $label=$day->format('o-W');
$label=$day->format('o \WW');
$label=$day->format('W o D');
$label=gmdate('o-W T e');
$label=gmdate('o W N');
$label=$day->format('o-\Y-W-o');
$label=gmdate('\m\d o-W');
