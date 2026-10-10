<?php
enum Choice {case Left;case Right;} function label(Choice $c) {return <error descr="Handle every reachable enum case.">match($c) {Choice::Left=>'L'}</error>;}
