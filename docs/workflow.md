# GitHub Actions Workflow

```mermaid
graph TD
    %% Define styles
    classDef triggerStyle fill:#ffebee,stroke:#c62828,stroke-width:2px;
    classDef versionStyle fill:#e8f5e9,stroke:#2e7d32,stroke-width:2px;
    classDef buildStyle fill:#e3f2fd,stroke:#1565c0,stroke-width:2px;
    classDef cleanupStyle fill:#fff3e0,stroke:#ef6c00,stroke-width:2px;
    classDef conditionStyle fill:#f3e5f5,stroke:#6a1b9a,stroke-width:1px;

    %% Trigger Events
    subgraph Triggers ["Trigger Events"]
        direction TB
        PR["Pull Request to main"]
        MergePR["PR Merge to main"]
        ManualTag["Manual Tag v*.*.*"]
    end

    %% Initial Version Actions
    PR --> VersionCheck["Version Check"]
    MergePR --> AutoPatch["Auto Patch Increment"]
    ManualTag --> MajorMinor["Major/Minor Version Change"]

    %% Version Management Flow
    subgraph VersionFlow["Version Management"]
        direction TB
        VersionCheck --> PreRelease["Pre-release: X.Y.Z-preN.M"]
        AutoPatch --> PatchRelease["Patch Version: X.Y.Z"]
        MajorMinor --> Release["Release Version: X.Y.Z"]
    end

    %% Build Workflows
    PreRelease & PatchRelease & Release --> DockerBuild

    subgraph DockerBuild["Docker Build Workflow"]
        direction TB
        Version["Get Release Version"] --> ReleaseBuild["Build And Push"]
        ReleaseBuild --> CleanupTrigger["Trigger Cleanup"]
    end

    %% Helm Chart Flow
    PreRelease & PatchRelease & Release -->  HelmBuild

    subgraph HelmBuild["Helm Build Workflow"]
        direction TB
        H_Version["Update Chart Version"] --> H_Package["Package Chart"]
        H_Package --> |"Release Only"| H_Publish["Publish Chart"]
    end

    %% Cleanup Workflow
    subgraph Cleanup["Cleanup Workflow"]
        direction TB
        C_Login["Docker Login"] --> C_List["List Old Releases"]
        C_List --> C_Delete["Delete Old Images & Tags"]
    end

    %% Workflow Dependencies
    CleanupTrigger --> Cleanup

    %% Apply styles
    class PR,MergePR,ManualTag triggerStyle;
    class VersionCheck,AutoPatch,MajorMinor,PreRelease,PatchRelease,Release versionStyle;
    class Version,CheckIfRelease,PRBuild,ReleaseBuild,H_Version,H_Package,H_Publish buildStyle;
    class C_Login,C_List,C_Delete,CleanupTrigger cleanupStyle;
    class PR_Cond,Tag_Cond,Manual conditionStyle;
```

## Workflow Description

### Triggers
- Pull requests to main branch (excluding paths: charts/**)
- Tag pushes matching v*.*.* pattern
- Manual workflow dispatch

### Docker Build Workflow
1. Checks if the trigger is from main branch or PR
2. Generates semantic version based on git history
   - For main branch: X.Y.Z
   - For PR: X.Y.Z-preN.M (where N is PR number and M is increment)
3. Builds Docker image
4. Pushes image only for releases (non-PR events)
5. Triggers cleanup workflow for old releases

### Cleanup Workflow
1. Authenticates with Docker registry using provided credentials
2. Lists releases older than specified retention period (default: 7 days)
3. Removes outdated images and tags to maintain registry cleanliness

### Version Management
- Release versions follow semantic versioning (X.Y.Z)
- Pre-release versions include PR number and increment (X.Y.Z-preN.M)
- Version tags are generated automatically based on git history
- Version format is consistent across Docker images and Helm charts

### Concurrency Control
- Workflows are grouped by workflow name and git ref
- In-progress workflows are cancelled when new ones are triggered
- Prevents redundant builds and resource waste