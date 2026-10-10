<?php
namespace Negative0 { strtok($a,',');strtok(','); }
namespace Negative1 { strtok($a,',');strtok($b,':'); }
namespace Negative2 { strtok($a,',');unknown();strtok($b,':');strtok(','); }
namespace Negative3 { strtok($a,',');strtok(',');strtok($b,':');strtok(','); }
