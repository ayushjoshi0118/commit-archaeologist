#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .setup(|_app| {
            // Setup hooks: can spawn Go sidecar process or verify local backend daemon
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running Commit Archaeologist desktop shell");
}
