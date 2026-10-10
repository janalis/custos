<?php
use function pcntl_async_signals as builtinCall12;
if (<warning descr="Query the signal state after changing it.">!builtinCall12(true)</warning>) { throw new RuntimeException(); }
