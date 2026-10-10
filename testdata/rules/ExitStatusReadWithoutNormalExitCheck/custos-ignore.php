<?php
// @custos-ignore ExitStatusReadWithoutNormalExitCheck
pcntl_waitpid($pid,$status); pcntl_wexitstatus($status);
