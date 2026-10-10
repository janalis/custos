<?php
use function pcntl_waitpid as builtinCall11;
builtinCall11($pid,$status); <warning descr="Check normal termination before decoding the exit code.">pcntl_wexitstatus($status)</warning>;
