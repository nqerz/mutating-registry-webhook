# GitHub Actions Workflow

```mermaid
graph TD
    %% Trigger Events
    PR[Pull Request to main] --> DockerBuild[Docker Build]
    PR --> HelmBuild[Helm Build]
    Tag[Push tag v*.*.* to main] --> DockerBuild
    Tag --> HelmBuild
    Manual[Manual Trigger] --> DockerBuild
    Manual --> HelmBuild

    %% Docker Build Workflow
    subgraph DockerBuild[Docker Build Workflow]
        Check[Check PR Merge] --> Version[Get Semantic Version]
        Version --> Build[Build & Push Image]
        Build --> CleanupTrigger[Trigger Cleanup]
    end

    %% Helm Build Workflow
    subgraph HelmBuild[Helm Build Workflow]
        H_Check[Check Chart Changes] --> H_Version[Update Chart Version]
        H_Version --> H_Package[Package Chart]
        H_Package --> H_Publish[Publish Chart]
    end

    %% Cleanup Workflow
    subgraph Cleanup[Cleanup Workflow]
        C_Login[Docker Login] --> C_List[List Old Releases]
        C_List --> C_Delete[Delete Old Images & Tags]
    end

    %% Workflow Dependencies
    CleanupTrigger --> Cleanup

    %% Conditions
    subgraph Conditions[Workflow Conditions]
        direction LR
        PR_Cond["Ignored: charts/**"] -.- PR
        Tag_Cond["Pattern: v*.*.* tags"] -.- Tag
        Helm_Cond["Changed: charts/**"] -.- HelmBuild
    end

    %% Version Management
    subgraph Versioning[Version Management]
        direction LR
        Main["Main Branch"] --> Release["Release Version
Format: X.Y.Z"]
        PR_Ver["PR Branch"] --> PreRelease["Pre-release Version
Format: X.Y.Z-preN.M"]
    end

    classDef default fill:#f9f,stroke:#333,stroke-width:2px;
    classDef subgraphStyle fill:#fff,stroke:#333,stroke-width:2px;
    classDef conditionStyle fill:#e1f5fe,stroke:#333,stroke-width:1px;
    class DockerBuild,Cleanup,Versioning,Conditions,HelmBuild subgraphStyle;
    class PR_Cond,Tag_Cond,Helm_Cond conditionStyle;
```

## Workflow Description

### Triggers
- Pull requests to main branch (paths: cmd/**, Dockerfile)
- Pull requests affecting Helm charts (paths: charts/**)
- Tag pushes matching v*.*.* pattern
- Manual workflow dispatch

### Docker Build Workflow
1. Checks if the trigger is from main branch or PR
2. Generates semantic version based on git history
   - For main branch: X.Y.Z
   - For PR: X.Y.Z-preN.M
3. Builds and pushes Docker image if conditions are met
4. Triggers cleanup workflow

### Helm Build Workflow
1. Validates changes in charts directory
2. Updates chart version based on semantic versioning
3. Packages Helm chart
4. Publishes chart to registry

### Cleanup Workflow
1. Authenticates with Docker registry
2. Lists old releases based on retention policy
3. Removes outdated images and tags

### Version Management
- Release versions follow semantic versioning (X.Y.Z)
- Pre-release versions include PR number and run number
- Version tags are generated automatically based on git history
- Helm chart versions are synchronized with application versions