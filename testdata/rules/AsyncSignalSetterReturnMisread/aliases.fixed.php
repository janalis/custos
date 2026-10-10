<?php
use function pcntl_async_signals as builtinCall12;
builtinCall12(true);
if (!builtinCall12()) { throw new RuntimeException(); }
