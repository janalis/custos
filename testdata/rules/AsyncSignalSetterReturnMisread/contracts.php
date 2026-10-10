<?php
if (<warning descr="Query the signal state after changing it.">!pcntl_async_signals(true)</warning>) { throw new RuntimeException(); }
