# Base image supports Nvidia CUDA but also runs on CPU
FROM nvidia/cuda:12.6.2-base-ubuntu22.04

USER root
ENV TORCH_HOME=/data/models
ENV OMP_NUM_THREADS=1

# Install required tools
RUN apt update && apt install -y --no-install-recommends \
    build-essential \
    ffmpeg \
    git \
    python3 \
    python3-dev \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

# Clone Demucs repo
RUN git clone --single-branch --branch main https://github.com/adefossez/demucs /lib/demucs
WORKDIR /lib/demucs
RUN git checkout b9ab48cad45976ba42b2ff17b229c071f0df9390

# Install dependencies
RUN python3 -m pip install -e . "torch<2" "torchaudio<2" "numpy<2" --no-cache-dir

# Copy pre-downloaded models into the image
COPY ./models /data/models

# Test it once to trigger validation
RUN python3 -m demucs -d cpu --model htdemucs --out /tmp/out /lib/demucs/test.mp3 || true
RUN rm -rf /tmp/out

# Set up volumes for input/output
VOLUME /data/input
VOLUME /data/output

ENTRYPOINT ["python3", "-m", "demucs"]
