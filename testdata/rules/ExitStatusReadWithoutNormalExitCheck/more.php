<?php
pcntl_wait($status);<warning descr="Check normal termination before decoding the exit code.">pcntl_wexitstatus($status)</warning>;
