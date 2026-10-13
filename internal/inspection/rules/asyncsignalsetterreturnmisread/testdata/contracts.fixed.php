<?php
pcntl_async_signals(true);
if (!pcntl_async_signals()) { throw new RuntimeException(); }
