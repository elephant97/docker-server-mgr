package request

type PortMapping struct {
	HostPort      string `json:"host_port"`      // 호스트 포트
	ContainerPort string `json:"container_port"` // 컨테이너 포트
}

type CreateRequest struct {
	UserId        int64         `json:"user_id"` // 사용자 ID
	Image         string        `json:"image"`
	Tag           string        `json:"tag"`            // Docker 이미지 태그
	ContainerName string        `json:"container_name"` // 컨테이너 이름
	Cmd           []string      `json:"cmds"`
	Ports         []PortMapping `json:"ports"`   // 포트 매핑 정보
	TTL           int           `json:"ttl"`     // 단위: 초
	Volumes       []string      `json:"volumes"` // 호스트 디렉토리 바인딩
}

type DeleteRequest struct {
	ContainerId string `json:"container_id"` // 컨테이너 ID
}
