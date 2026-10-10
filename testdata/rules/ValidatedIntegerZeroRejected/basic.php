<?php
if (!<warning descr="Compare integer validation failure strictly with false.">filter_var($input, FILTER_VALIDATE_INT)</warning>) { return false; }
