<?php
try { work(); } catch(Throwable $e) {} catch(<warning descr="Move the specific catch before the broader catch.">RuntimeException</warning> $e) {}
