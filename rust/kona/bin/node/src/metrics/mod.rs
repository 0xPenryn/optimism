//! Global metrics for `kona-node`

mod cli_opts;
pub use cli_opts::{init_rollup_config_metrics, CliMetrics};

mod version;
pub use version::VersionInfo;
